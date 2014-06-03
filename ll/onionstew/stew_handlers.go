package onionstew

import (
	"io"
	"net"

	"github.com/KirisurfProject/kilog"
)

func (ctx *stew_ctx) attacht_client(client io.ReadWriteCloser, tgt string) {
	defer client.Close()

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
	xaxa := (remconn).(*net.TCPConn)
	xaxa.SetLinger(0)
	xaxa.SetNoDelay(true)

	if err != nil {
		kilog.Debug("attacht_remote failed to connect to %d!", remote_addr)
		return
	}

	defer xaxa.Close()

	tunnel_connection(ctx, connid, xaxa)
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

	this_close := func() {
		select {
		case <-local_close:
		default:
			close(local_close)
		}
	}

	defer func() {
		select {
		case <-ctx.killswitch:
		case ctx.write_ch <- close_message:
		}
		this_close()
		ctx.conntable[connid] = nil
	}()

	// Semaphore for stoppings of sendings after 256
	stop_sem := make(chan bool, 256)

	for i := 0; i < 256; i++ {
		select {
		case stop_sem <- true:
		default:
		}
	}

	// Read from read_ch
	go func() {
		datactr := 256
		for {
			select {
			case <-ctx.killswitch:
				socket.Close()
				return
			case newpkt := <-read_ch:
				if newpkt.category == m_close {
					kilog.Debug("Closing an encapsulated socket...")
					socket.Close()
					kilog.Debug("Close() returned")
					this_close()
					return
				}
				if newpkt.category == m_data {
					_, err := socket.Write(newpkt.payload)
					if err != nil {
						socket.Close()
						return
					}
					datactr--
					if datactr == 0 {
						kilog.Debug("Bucket drained, sending m_more")
						datactr = 256
						ctx.write_ch <- stew_message{m_more, connid, []byte("")}
					}
				}
				if newpkt.category == m_more {
					kilog.Debug("Got m_more!")
					for i := 0; i < 256; i++ {
						select {
						case stop_sem <- true:
						default:
						}
					}
				}
			case <-local_close:
				socket.Close()
				return
			}
		}
	}()

	buff := make([]byte, 16384)
	// Read from socket
	for {
		select {
		case <-ctx.killswitch:
			return
		case <-local_close:
			kilog.Debug("Socket closed.")
			return
		default:
			n, err := socket.Read(buff)
			if err != nil {
				kilog.Debug("Socket closed. (%s)", err.Error())
				return
			}
			thing := make([]byte, n)
			copy(thing, buff)
			msg := stew_message{m_data, connid, thing}
			// Sem dec
			select {
			case <-stop_sem:
			case <-ctx.killswitch:
				return
			default:
				kilog.Debug("Waiting for m_more...")
				select {
				case <-stop_sem:
				case <-ctx.killswitch:
					return
				case <-local_close:
					kilog.Debug("Socket closed.")
					return
				}
			}
			select {
			case <-ctx.killswitch:
				return
			case ctx.write_ch <- msg:
			case <-local_close:
				kilog.Debug("Socket closed.")
				return
			}
		}
	}
}
