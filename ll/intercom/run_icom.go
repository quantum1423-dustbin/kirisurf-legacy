package intercom

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/KirisurfProject/kilog"
)

func run_icom_ctx(ctx *icom_ctx, KILL func(), is_server bool) {
	defer KILL()
	socket_table := make([]chan icom_msg, 65536)
	stable_lock := make(chan bool, 1)
	stable_lock <- true

	prob_dist := MakeProbDistro()
	junk_chan := make(chan bool, 1024)

	// Write junk echo packets to mask webpage loading
	go func() {
		defer KILL()
		for {
			desired_size := prob_dist.Draw()
			select {
			case <-ctx.killswitch:
				return
			case <-junk_chan:
				select {
				case <-ctx.killswitch:
				case ctx.write_ch <- icom_msg{icom_ignore,
					0, make([]byte, desired_size)}:
				}
			}
		}
	}()

	// Write packets
	go func() {
		defer KILL()
		for {
			select {
			case <-ctx.killswitch:
				return
			case xaxa := <-ctx.write_ch:
				buffer := new(bytes.Buffer)
				desired_size := prob_dist.Draw()
				prob_dist.Juggle()
				xaxa.WriteTo(buffer)
				if desired_size > len(xaxa.body) {
					excess := desired_size - len(xaxa.body)
					padd := icom_msg{icom_ignore, 0, make([]byte, excess)}
					padd.WriteTo(buffer)
				}
				if xaxa.flag == icom_data {
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
	go func() {
		for {
			select {
			case <-ctx.killswitch:
				return
			case <-time.After(time.Second * time.Duration(rand.Int()%10)):
				select {
				case <-ctx.killswitch:
					return
				case ctx.write_ch <- icom_msg{icom_ignore, 0, make([]byte, 0)}:
				}
			}
		}
	}()

	// Client side. Writes stuff.
	if !is_server {
		defer KILL()
		go func() {
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
				for i := 0; i < 65536; i++ {
					if socket_table[i] == nil {
						connid = i
						break
					}
				}
				ctx.write_ch <- icom_msg{icom_open, connid, make([]byte, 0)}
				xaxa := make(chan icom_msg, 2048)
				socket_table[connid] = xaxa
				fmt.Println("Client side tunneling connid", connid)
				stable_lock <- true
				go func() {
					icom_tunnel(ctx, KILL, incoming, connid, xaxa)
					fmt.Println("Freeing connid %d", connid)
					<-stable_lock
					socket_table[connid] = nil
					stable_lock <- true
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

		// Now work with the packet
		if justread.flag == icom_ignore {
			continue
		}
		if justread.flag == icom_open && is_server {
			// Open a connection! The caller of accept will unblock this call.
			conn, err := VSConnect(ctx.our_srv)
			if err != nil {
				return
			}
			xaxa := make(chan icom_msg, 2048)
			<-stable_lock
			socket_table[justread.connid] = xaxa
			stable_lock <- true
			// Tunnel the connection
			fmt.Println("Server side tunneling connid", justread.connid)
			go icom_tunnel(ctx, KILL, conn, justread.connid, xaxa)
		} else if justread.flag == icom_data ||
			justread.flag == icom_more {
			<-stable_lock
			if socket_table[justread.connid] == nil {
				kilog.Debug("Tried to send packet to nonexistent connid!\n%s", string(justread.body))
				stable_lock <- true
				return
			}
			ch := socket_table[justread.connid]
			stable_lock <- true
			// Forward the data to the socket
			select {
			case ch <- justread:
			case <-ctx.killswitch:
				return
			default:
				fmt.Println("Blocked on forward!")
			}
		} else if justread.flag == icom_close {
			<-stable_lock
			if socket_table[justread.connid] == nil {
				kilog.Debug("Tried to send packet to nonexistent connid!\n%v", justread)
				continue
			}
			ch := socket_table[justread.connid]
			socket_table[justread.connid] = nil
			stable_lock <- true
			select {
			case ch <- justread:
			case <-ctx.killswitch:
				return
			}
		} else {
			kilog.Debug("** icom_ctx dead ** due to invalid packet")
			return
		}
	}
}
