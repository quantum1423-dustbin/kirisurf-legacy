// subcircuit_server.go
package main

import (
	"encoding/gob"
	"errors"
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
	"net"

	"github.com/coreos/go-log/log"
)

func sc_server_handler(wire net.Conn) error {
	defer wire.Close()
	owire, err := kiss.Kiriobfs_handshake_server(wire)
	log.Debug("Of dones for obfs layer")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	awire, err := kiss.KiSS_handshake_server(owire, MasterKey)
	log.Debug("Of dones in kiss layer")
	if err != nil {
		log.Error(err.Error())
		return err
	}
	// Now awire is the wire
	gobreader := gob.NewDecoder(awire)
	var cmd sc_message
	err = gobreader.Decode(&cmd)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Debug(cmd)
	if cmd.Msg_type == SC_EXTEND {
		theirnode := dirclient.PKeyLookup(cmd.Msg_arg)
		if theirnode == nil {
			return errors.New("Watif")
		}
		actwire, err := net.Dial("tcp", theirnode.Address)
		if err != nil {
			return err
		}
		remwire, err := kiss.Kiriobfs_handshake_client(actwire)
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
					log.Error(err.Error())
					client.Close()
					continue
				}
				go func() {
					err := sc_server_handler(client)
					if err != nil {
						log.Error(err.Error())
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
