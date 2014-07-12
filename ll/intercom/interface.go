package intercom

import (
	"errors"
	"io"
	"kirisurf/ll/kiss"
	"net"
	"strings"
	"sync"
)

type IntercomServer chan io.ReadWriteCloser

func (server IntercomServer) Accept() io.ReadWriteCloser {
	return <-server
}

func MakeIntercomServer(laddr string) IntercomServer {
	bla := strings.Split(laddr, "@")

	toret := make(chan io.ReadWriteCloser)
	listener, err := net.Listen("tcp", bla[1])
	if err != nil {
		panic(err.Error())
	}
	go func() {
		for {
			newclient, err := listener.Accept()
			if err != nil {
				continue
			}
			realclient, err := kiss.Obfs4fHandshake(newclient, true, bla[0])
			if err != nil {
				panic(err.Error())
			}
			go func() {
				ctx := make_icom_ctx(realclient, true, true)
				go func() {
					defer realclient.Close()
					for {
						thing, err := ctx.our_srv.Accept()
						if err != nil {
							return
						}
						toret <- thing
					}
				}()
			}()
		}
	}()
	return IntercomServer(toret)
}

type IntercomDialer struct {
	mapping map[string]*icom_ctx
	lock    sync.Mutex
}

func MakeIntercomDialer() *IntercomDialer {
	xaxa := new(IntercomDialer)
	xaxa.mapping = make(map[string]*icom_ctx)
	return xaxa
}

func (dialer *IntercomDialer) Dial(host string) (io.ReadWriteCloser, error) {
	dialer.lock.Lock()
	defer dialer.lock.Unlock()

	bla := strings.Split(host, "@")
	host = bla[1]
	key := bla[0]

	// Establish the connection if doesn't exist
	if dialer.mapping[host] == nil {
		xaxa, err := net.Dial("tcp", host)
		if err != nil {
			return nil, err
		}
		really, err := kiss.Obfs4fHandshake(xaxa, false, key)
		if err != nil {
			return nil, err
		}
		dialer.mapping[host] = make_icom_ctx(really, false, true)
	}
	for i := 0; i < 10; i++ {
		select {
		case <-dialer.mapping[host].killswitch:
			xaxa, err := net.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			really, err := kiss.Obfs4fHandshake(xaxa, false, key)
			if err != nil {
				return nil, err
			}
			dialer.mapping[host] = make_icom_ctx(really, false, true)
		default:
		}
	}
	select {
	case <-dialer.mapping[host].killswitch:
		return nil, errors.New("Connection keeps on failing.")
	default:
		return VSConnect(dialer.mapping[host].our_srv)
	}
}
