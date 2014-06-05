package intercom

import (
	"io"

	"github.com/KirisurfProject/kilog"
)

func run_icom_ctx(ctx *icom_ctx, KILL func()) {
	defer KILL()
	socket_table := make(map[int]chan icom_msg)

	// Downstream link
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
		if justread.flag == icom_open {
			// Open a connection! The caller of accept will unblock this call.
			xaxa := make(chan icom_msg, 256)
			socket_table[justread.connid] = xaxa
			// Tunnel the connection
			go icom_tunnel(ctx, xaxa)
		} else if justread.flag == icom_data ||
			justread.flag == icom_more {
			if socket_table[justread.connid] == nil {
				kilog.Debug("Tried to send packet to nonexistent connid!")
				return
			}
			// Forward the data to the socket
			socket_table[justread.connid] <- justread
		} else if justread.flag == icom_close {
			if socket_table[justread.connid] == nil {
				kilog.Debug("Tried to send packet to nonexistent connid!")
				return
			}
			ch := socket_table[justread.connid]
			delete(socket_table, justread.connid)
			ch <- justread
		} else {
			kilog.Debug("** icom_ctx dead ** due to invalid packet")
			return
		}
	}
}
