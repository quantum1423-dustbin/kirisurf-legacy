package intercom

import (
	"io"
	"kirisurf/ll/socks5"
	"net"
	"time"

	"github.com/KirisurfProject/kilog"
	"github.com/KirisurfProject/rwcutils"
)

type MultiplexClient struct {
	*icom_ctx
}

func MakeMultiplexClient(transport io.ReadWriteCloser) MultiplexClient {
	return MultiplexClient{make_icom_ctx(transport, false, false)}
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
	ctx := make_icom_ctx(transport, true, false)
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
			rlrem := rwcutils.RateLimit(remote, 40, 200)
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
