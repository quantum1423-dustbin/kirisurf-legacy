package intercom

import (
	"bytes"
	"math/rand"
	"time"

	"github.com/KirisurfProject/kilog"
)

func run_icom_ctx(ctx *icom_ctx, KILL func(), is_server bool, do_junk bool, PAUSELIM int) {
	defer KILL()
	socket_table := make([]chan icom_msg, 65536)
	stable_lock := make(chan bool, 1)
	stable_lock <- true

	prob_dist := MakeProbDistro()
	junk_chan := make(chan bool)

	// Write junk echo packets to mask webpage loading
	if do_junk {
		go func() {
			defer KILL()
			for {
				desired_size := prob_dist.Draw() * 2
				select {
				case <-ctx.killswitch:
					return
				case <-junk_chan:
					select {
					case <-ctx.killswitch:
					case ctx.write_ch <- icom_msg{icom_ignore,
						0, make([]byte, desired_size)}:
					default:
					}
				}
			}
		}()
	}

	// Write packets
	go func() {
		defer KILL()
		for {
			select {
			case <-ctx.killswitch:
				return
			case xaxa := <-ctx.write_ch:
				lel := "data"
				if xaxa.flag == icom_ignore {
					lel = "junk"
				} else if xaxa.flag == icom_open {
					lel = "open"
				} else if xaxa.flag == icom_close {
					lel = "clos"
				} else if xaxa.flag == icom_more {
					lel = "more"
				}

				kilog.FineDebug("[ICOM] -> %v\t%v\t%v:%v", do_junk, xaxa.connid,
					lel, len(xaxa.body))

				buffer := new(bytes.Buffer)
				desired_size := prob_dist.Draw()
				prob_dist.Juggle()
				xaxa.WriteTo(buffer)
				if desired_size > len(xaxa.body) && do_junk {
					excess := desired_size - len(xaxa.body)
					padd := icom_msg{icom_ignore, 0, make([]byte, excess)}
					padd.WriteTo(buffer)
				}
				if xaxa.flag == icom_data && do_junk {
					// Draw a waiting period
					wsecs := rand.ExpFloat64() * 3
					wms := int64(wsecs * 1000)
					// Spin off a goroutine to do this!
					go func() {
						time.Sleep(time.Millisecond * time.Duration(wms))
						select {
						case junk_chan <- true:
						default:
						}
					}()
				}
				_, err := ctx.underlying.Write(buffer.Bytes())
				if err != nil {
					return
				}
			}
		}
	}()

	// Keepalive pakkets
	if do_junk {
		go func() {
			for {
				select {
				case <-ctx.killswitch:
					return
				case <-time.After(time.Second * time.Duration(rand.Int()%5)):
					select {
					case <-ctx.killswitch:
						return
					case ctx.write_ch <- icom_msg{icom_ignore, 0, make([]byte, 0)}:
					}
				}
			}
		}()
	}

	// Client side. Writes stuff.
	if !is_server {
		go func() {
			defer KILL()
			for {
				// Accepts clients
				incoming, err := ctx.our_srv.Accept()
				if err != nil {
					kilog.Debug("** icom_ctx dead @ client accept **")
					return
				}
				// Find a connid
				<-stable_lock
				connid := 0
				startidx := rand.Int() % 65536

				// DEBUG!!!
				//startidx = 0

				for i := 0; i < 65536; i++ {
					if socket_table[(i+startidx)%65536] == nil {
						connid = (i + startidx) % 65536
						break
					}
				}
				ctx.write_ch <- icom_msg{icom_open, connid, make([]byte, 0)}
				xaxa := make(chan icom_msg, PAUSELIM)
				socket_table[connid] = xaxa
				stable_lock <- true
				go func() {
					if !do_junk {
						kilog.Debug("ICOM: Opened connid %d", connid)
					}
					icom_tunnel(ctx, KILL, incoming, connid, xaxa, do_junk, PAUSELIM)
					<-stable_lock
					socket_table[connid] = nil
					stable_lock <- true
					if !do_junk {
						kilog.Debug("ICOM: Closed connid %d", connid)
					}
				}()
			}
		}()
	}

	// Reading link
	for {
		var justread icom_msg
		err := justread.ReadFrom(ctx.underlying)
		if err != nil {
			kilog.Debug("** icom_ctx dead @ body ** due to %s", err.Error())
			return
		}

		xaxa := justread

		lel := "data"
		if xaxa.flag == icom_ignore {
			lel = "junk"
		} else if xaxa.flag == icom_open {
			lel = "open"
		} else if xaxa.flag == icom_close {
			lel = "clos"
		} else if xaxa.flag == icom_more {
			lel = "more"
		}

		kilog.FineDebug("[ICOM] <- %v\t%v\t%v:%v", do_junk, xaxa.connid, lel, len(xaxa.body))

		// Now work with the packet
		if justread.flag == icom_ignore {
			continue
		}
		if justread.flag == icom_open && is_server {
			if !do_junk {
				kilog.Debug("ICOM: Accepted connid %d", justread.connid)
			}
			// Open a connection! The caller of accept will unblock this call.
			conn, err := VSConnect(ctx.our_srv)
			if err != nil {
				return
			}
			xaxa := make(chan icom_msg, PAUSELIM)
			<-stable_lock
			socket_table[justread.connid] = xaxa
			stable_lock <- true
			go func() {
				kilog.Debug("ICOM: Began processing connid %d", justread.connid)
				// Tunnel the connection
				icom_tunnel(ctx, KILL, conn, justread.connid, xaxa, do_junk, PAUSELIM)
				if !do_junk {
					kilog.Debug("ICOM: Closed connid %d", justread.connid)
				}
				<-stable_lock
				socket_table[justread.connid] = nil
				stable_lock <- true
			}()
		} else if justread.flag == icom_data ||
			justread.flag == icom_more ||
			justread.flag == icom_close {
			<-stable_lock
			if socket_table[justread.connid] == nil {
				stable_lock <- true
				continue
			}
			ch := socket_table[justread.connid]
			stable_lock <- true
			// Forward the data to the socket
			select {
			case ch <- justread:
			case <-ctx.killswitch:
				return
			default:
			}
		} else {
			kilog.Debug("** icom_ctx dead ** due to invalid packet")
			return
		}
	}
}
