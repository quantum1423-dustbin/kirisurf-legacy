// subcircuit_server.go
package main

import (
	"errors"
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
	"net"

	"github.com/KirisurfProject/kilog"
	"github.com/coreos/go-log/log"
)

func sc_server_handler(wire net.Conn) error {
	defer wire.Close()
	owire, err := kiss.Obfs3fHandshake(wire, true)
	if err != nil {
		//log.Error(err.Error())
		return err
	}
	awire, err := kiss.TransportHandshake(MasterKey, owire,
		func([]byte) bool { return true })
	if err != nil {
		kilog.Debug(err.Error())
		return err
	}
	// Now awire is the wire
	cmd, err := read_sc_message(awire)
	if err != nil {
		WARNING(err.Error())
		return err
	}
	if cmd.Msg_type == SC_EXTEND {
		theirnode := dirclient.PKeyLookup(cmd.Msg_arg)
		if theirnode == nil {
			return errors.New("Watif")
		}
		actwire, err := net.Dial("tcp", theirnode.Address)
		if err != nil {
			return err
		}
		remwire, err := kiss.Obfs3fHandshake(actwire, false)
		if err != nil {
			return err
		}
		go func() {
			io.Copy(remwire, awire)
			remwire.Close()
		}()
		io.Copy(awire, remwire)
		awire.Close()
	} else if cmd.Msg_type == SC_TERMINATE && MasterConfig.General.IsExit {
		kilog.Debug("SC_TERMINATE received")
		e2e_server_handler(awire)
	}
	return nil
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
				log.Debug("Of acceptings client: %s", client.RemoteAddr())
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
