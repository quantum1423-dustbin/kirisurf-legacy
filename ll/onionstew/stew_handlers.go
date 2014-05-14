package onionstew

import (
	"io"
	"net"

	"github.com/KirisurfProject/kilog"
)

func (ctx *stew_ctx) attacht_client(client io.ReadWriteCloser) {
	defer client.Close()

	// Read the socks first
	tgt, err := socks5_handshake(client)
	if err != nil {
		return
	}

	kilog.Debug("SOCKS5 handshake done to %s", tgt)

	// We need to find an appropriate channel slot first!
	connid := <-ctx.number_ch
	defer func() {
		client.Close()
		ctx.number_ch <- connid
	}()
	read_ch := make(chan stew_message, 256)
	ctx.lock.Lock()
	ctx.conntable[connid] = read_ch
	ctx.lock.Unlock()

	// Send connection request
	select {
	case ctx.write_ch <- stew_message{m_open, connid, []byte(tgt)}:
	case <-ctx.killswitch:
		return
	}

	tunnel_connection(ctx, connid, client)
}

func (ctx *stew_ctx) attacht_remote(remote_addr string, connid int) {
	remconn, err := net.Dial("tcp", remote_addr)
	if err != nil {
		kilog.Debug("attacht_remote failed to connect to %d!", remote_addr)
		return
	}

	defer remconn.Close()

	tunnel_connection(ctx, connid, remconn)
}

func tunnel_connection(ctx *stew_ctx, connid int, socket io.ReadWriteCloser) {
	local_close := make(chan bool)
	ctx.lock.RLock()
	read_ch := ctx.conntable[connid]
	ctx.lock.RUnlock()

	// Catches deadlocks: destroy client when killswitch received
	go func() {
		select {
		case <-ctx.killswitch:
		case <-local_close:
		}
		socket.Close()
	}()

	close_message := stew_message{m_close, connid, []byte("")}
	defer func() {
		select {
		case <-ctx.killswitch:
		case ctx.write_ch <- close_message:
		}
		close(local_close)
		ctx.conntable[connid] = nil
	}()

	// Read from read_ch
	go func() {
		for {
			select {
			case <-ctx.killswitch:
				socket.Close()
				return
			case newpkt := <-read_ch:
				if newpkt.category == m_close {
					socket.Close()
					return
				}
				if newpkt.category == m_data {
					_, err := socket.Write(newpkt.payload)
					if err != nil {
						socket.Close()
						return
					}
				}
			case <-local_close:
				socket.Close()
				return
			}
		}
	}()
	buff := make([]byte, 4096)
	// Read from socket
	for {
		select {
		case <-ctx.killswitch:
			return
		default:
			n, err := socket.Read(buff)
			if err != nil {
				return
			}
			thing := make([]byte, n)
			copy(thing, buff)
			msg := stew_message{m_data, connid, thing}
			select {
			case <-ctx.killswitch:
				return
			case ctx.write_ch <- msg:
			}
		}
	}
}
