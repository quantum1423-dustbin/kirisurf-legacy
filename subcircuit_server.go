// subcircuit_server.go
package main

import (
	"fmt"
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/intercom"
	"kirisurf/ll/kiss"
	"net"
	"strconv"
	"strings"

	"github.com/KirisurfProject/kilog"
)

func sc_server_real_handler(_wire io.ReadWriteCloser) (err error) {
	wire, err := kiss.TransportHandshake(MasterKey, _wire,
		func([]byte) bool { return true })
	if err != nil {
		kilog.Debug("failed the transport handshake")
		return err
	}
	thing := make([]byte, 1)
	_, err = io.ReadFull(wire, thing)
	if err != nil {
		kilog.Debug("failed the readfull")
		return err
	}
	if thing[0] == 0 {
		// Terminate
		if !MasterConfig.General.IsExit {
			return nil
		}
		kilog.Debug("terminating")
		e2e_server_handler_old(wire)
	} else if thing[0] == 255 {
		// Terminate with NEW GENERATION
		if !MasterConfig.General.IsExit {
			return nil
		}
		kilog.Debug("terminating NG")
		e2e_server_handler(wire)
	} else {
		xaxa := make([]byte, thing[0])
		_, err := io.ReadFull(wire, xaxa)
		if err != nil {
			return err
		}
		key := string(xaxa)
		qqq := dirclient.PKeyLookup(key)
		if qqq == nil {
			kilog.Debug("Cannot find %s in %v", xaxa, dirclient.KDirectory)
			dirclient.RefreshDirectory()
			qqq = dirclient.PKeyLookup(key)
			if qqq == nil {
				return nil
			}
		}
		kilog.Debug("Continuing to %s", qqq.Address)
		remm, err := dialer.Dial(old2new(qqq.Address))
		if err != nil {
			return err
		}
		go func() {
			io.Copy(wire, remm)
			wire.Close()
		}()
		io.Copy(remm, wire)
		remm.Close()
	}
	return io.EOF
}

func sc_server_handler(_wire net.Conn) (err error) {
	defer func() {
		if err != nil {
			kilog.Debug("sc_server_handler returning err=%s", err.Error())
		}
	}()
	defer _wire.Close()
	wire, err := kiss.Obfs3fHandshake(_wire, true)
	if err != nil {
		//kilog.Debug(err.Error())
		return nil
	}
	return sc_server_real_handler(wire)
}

type SCServer struct {
	listener net.Listener
	killer   chan bool
}

func NewSCServer(addr string) SCServer {
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		panic(err.Error())
	}
	killer := make(chan bool)
	go func() {
		for {
			select {
			case <-killer:
				listener.Close()
				return
			default:
				// establish connection
				client, err := listener.Accept()
				//log.Debug("Of acceptings client: %s", client.RemoteAddr())
				if err != nil {
					CRITICAL(err.Error())
					client.Close()
					continue
				}
				go func() {
					err := sc_server_handler(client)
					if err != nil {
						//log.Error(err.Error())
					}
				}()
			}
		}
	}()
	return SCServer{listener, killer}
}

func (thing SCServer) Kill() {
	thing.killer <- true
}

func RegisterNGSCServer(addr string) {
	port, _ := strconv.Atoi(strings.Split(addr, ":")[1])
	naddr := fmt.Sprintf("kirisurf@%s:%d", strings.Split(addr, ":")[0], port+1)
	listener := intercom.MakeIntercomServer(naddr)
	go func() {
		for {
			nooclient := listener.Accept()
			go func() {
				sc_server_real_handler(nooclient)
			}()
		}
	}()
}
