package intercom

import (
	"io"
	"kirisurf/ll/socks5"
	"net"
	"time"

	"github.com/KirisurfProject/kilog"
)

type MultiplexClient struct {
	*icom_ctx
}

func MakeMultiplexClient(transport io.ReadWriteCloser) MultiplexClient {
	return MultiplexClient{make_icom_ctx(transport, false, false, 128)}
}

func (ctx MultiplexClient) SocksAccept(client io.ReadWriteCloser) (io.ReadWriteCloser, error) {
	return VSConnect(ctx.our_srv)
}

func (ctx *MultiplexClient) Alive() bool {
	select {
	case <-ctx.killswitch:
		return false
	default:
		return true
	}
}

func RunMultiplexServer(transport io.ReadWriteCloser) {
	ctx := make_icom_ctx(transport, true, false, 128)
	for {
		thing, err := ctx.our_srv.Accept()
		if err != nil {
			return
		}
		go func() {
			defer thing.Close()
			init_done := make(chan bool)
			go func() {
				select {
				case <-init_done:
					kilog.Debug("ICOM: Initialization done")
					return
				case <-time.After(time.Second * 10):
					kilog.Warning("ICOM: ** Client still no request after 10 secs **")
				}
			}()
			lenbts := make([]byte, 2)
			_, err := io.ReadFull(thing, lenbts)
			if err != nil {
				kilog.Debug("** Reading destination length failed! **")
				return
			}
			addr := make([]byte, int(lenbts[0])+int(lenbts[1])*256)
			_, err = io.ReadFull(thing, addr)
			if err != nil {
				kilog.Debug("** Reading destination failed! **")
				return
			}
			init_done <- true
			if addr[0] == 't' {
				addr = addr[1:]
			} else {
				kilog.Warning("UDP support not implemented yet!")
				thing.Write([]byte("NOIM"))
				return
			}

			remote, err := net.DialTimeout("tcp", string(addr), time.Second*20)
			if err != nil {
				kilog.Debug("Connection to %s failed: %s", addr, err.Error())
				e := err.(net.Error)
				if e.Timeout() {
					thing.Write([]byte("TMOT"))
				} else {
					thing.Write([]byte("FAIL"))
				}
				return
			}
			defer remote.Close()
			rlrem := remote
			go func() {
				defer rlrem.Close()
				io.Copy(rlrem, thing)
			}()
			kilog.Debug("Opened connection to %s", addr)
			thing.Write([]byte("OKAY"))
			io.Copy(thing, rlrem)
		}()
	}
}

func RunMultiplexSOCKSServer(transport io.ReadWriteCloser) {
	ctx := make_icom_ctx(transport, true, false, 2048)
	for {
		thing, err := ctx.our_srv.Accept()
		if err != nil {
			return
		}
		go func() {
			defer thing.Close()
			addr, err := socks5.ReadRequest(thing)
			if err != nil {
				return
			}
			remote, err := net.DialTimeout("tcp", addr, time.Second*20)
			if err != nil {
				kilog.Debug("Connection to %s failed: %s", addr, err.Error())
				e := err.(net.Error)
				if e.Timeout() {
					socks5.CompleteRequest(0x06, thing)
				} else {
					socks5.CompleteRequest(0x01, thing)
				}
				return
			}
			defer remote.Close()
			rlrem := remote
			err = socks5.CompleteRequest(0x00, thing)
			if err != nil {
				return
			}
			go func() {
				defer rlrem.Close()
				io.Copy(rlrem, thing)
			}()
			kilog.Debug("Opened connection to %s", addr)
			io.Copy(thing, rlrem)
		}()
	}
}
