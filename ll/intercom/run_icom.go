package intercom

import (
	"encoding/binary"
	"io"

	"github.com/KirisurfProject/kilog"
)

func run_icom_ctx(ctx *icom_ctx, KILL func(), is_server bool) {
	defer KILL()
	socket_table := make(map[int]chan icom_msg)

	// Write packets
	go func() {
		defer KILL()
		for {
			select {
			case <-ctx.killswitch:
				return
			case xaxa := <-ctx.write_ch:
				towrite := make([]byte, 5+len(xaxa.body))
				copy(towrite[5:], xaxa.body)
				towrite[0] = byte(xaxa.flag)
				binary.BigEndian.PutUint16(towrite[1:3], uint16(xaxa.connid))
				binary.BigEndian.PutUint16(towrite[3:5], uint16(len(xaxa.body)))
				_, err := ctx.underlying.Write(towrite)
				if err != nil {
					kilog.Debug("** icom_ctx dead @ write ** due to %s", err.Error())
					return
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
				}
				//
			}
		}()
	}

	// Reading link
	metabuf := make([]byte, 5)
	for {
		// Read metadata of next pkt
		_, err := io.ReadFull(ctx.underlying, metabuf)
		if err != nil {
			kilog.Debug("** icom_ctx dead @ metadata ** due to %s", err.Error())
			return
		}
		// Parse md
		var justread icom_msg
		justread.flag = int(metabuf[0])
		justread.connid = int(metabuf[1])*256 + int(metabuf[2])
		length := int(metabuf[3])*256 + int(metabuf[4])
		justread.body = make([]byte, length)
		// Read the body
		_, err = io.ReadFull(ctx.underlying, justread.body)
		if err != nil {
			kilog.Debug("** icom_ctx dead @ body ** due to %s", err.Error())
			return
		}

		// Now work with the packet
		if justread.flag == icom_open && is_server {
			// Open a connection! The caller of accept will unblock this call.
			xaxa := make(chan icom_msg, 256)
			socket_table[justread.connid] = xaxa
			// Tunnel the connection
			go icom_tunnel(ctx, xaxa, justread.connid)
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
			}
		} else if justread.flag == icom_close {
			if socket_table[justread.connid] == nil {
				kilog.Debug("Tried to send packet to nonexistent connid!")
				return
			}
			ch := socket_table[justread.connid]
			delete(socket_table, justread.connid)
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
