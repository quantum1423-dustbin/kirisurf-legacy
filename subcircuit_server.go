// subcircuit_server.go
package main

import (
	"encoding/gob"
	"libkiss"
	"net"

	"github.com/coreos/go-log/log"
)

func sc_server_handler(wire net.Conn) error {
	defer wire.Close()
	owire, err := libkiss.Kiriobfs_handshake_server(wire)
	log.Debug("Of dones for obfs layer")
	if err != nil {
		return err
	}
	awire, err := libkiss.KiSS_handshake_server(owire, MasterKey)
	log.Debug("Of dones in kiss layer")
	if err != nil {
		return err
	}
	// Now awire is the wire
	gobreader := gob.NewDecoder(awire)
	var cmd sc_message
	err = gobreader.Decode(&cmd)
	if err != nil {
		return err
	}
	log.Debug(cmd)
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
