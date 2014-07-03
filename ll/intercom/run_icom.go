package intercom

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/KirisurfProject/kilog"
)

func run_icom_ctx(ctx *icom_ctx, KILL func(), is_server bool) {
	defer KILL()
	socket_table := make([]chan icom_msg, 65536)

	prob_dist := MakeProbDistro()

	// Write packets
	go func() {
		defer KILL()
		for {
			select {
			case <-ctx.killswitch:
				return
			case xaxa := <-ctx.write_ch:
				desired_size := prob_dist.Draw()
				prob_dist.Juggle()
				err := xaxa.WriteTo(ctx.underlying)
				if err != nil {
					kilog.Debug("** icom_ctx dead @ write ** due to %s", err.Error())
					return
				}
				if desired_size > len(xaxa.body) {
					excess := desired_size - len(xaxa.body)
					padd := icom_msg{icom_ignore, 0, make([]byte, excess)}
					err := padd.WriteTo(ctx.underlying)
					if err != nil {
						kilog.Debug("** icom_ctx dead @ write ** due to %s", err.Error())
						return
					}
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
				go icom_tunnel(ctx, KILL, incoming, connid, xaxa)
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
			conn := VSConnect(ctx.our_srv)
			xaxa := make(chan icom_msg, 2048)
			socket_table[justread.connid] = xaxa
			// Tunnel the connection
			fmt.Println("Server side tunneling connid", justread.connid)
			go icom_tunnel(ctx, KILL, conn, justread.connid, xaxa)
		} else if justread.flag == icom_data ||
			justread.flag == icom_more {
			if socket_table[justread.connid] == nil {
				kilog.Debug("Tried to send packet to nonexistent connid!")
				return
			}
			// Forward the data to the socket
			select {
			case socket_table[justread.connid] <- justread:
			case <-ctx.killswitch:
				return
			default:
				fmt.Println("Blocked on forward!")
			}
		} else if justread.flag == icom_close {
			if socket_table[justread.connid] == nil {
				kilog.Debug("Tried to send packet to nonexistent connid!")
				return
			}
			ch := socket_table[justread.connid]
			socket_table[justread.connid] = nil
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
