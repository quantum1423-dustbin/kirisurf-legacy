package intercom

import (
	"io"
	"kirisurf/ll/socks5"
	"net"
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
			remote, err := net.Dial("tcp", addr)
			if err != nil {
				return
			}
			defer remote.Close()
			err = socks5.CompleteRequest(thing)
			if err != nil {
				return
			}
			go func() {
				defer remote.Close()
				io.Copy(remote, thing)
			}()
			io.Copy(thing, remote)
		}()
	}
}
