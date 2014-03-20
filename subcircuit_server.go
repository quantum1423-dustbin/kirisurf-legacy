// subcircuit_server.go
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"libkiridir"
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
	// Now listen to commands
	command := make([]byte, 5)
	_, err = io.ReadFull(awire, command)
	if err != nil {
		return err
	}
	if string(command) == "CONN " {
		// Now read line as the next pubkey
		dasbuf := bufio.NewReader(awire)
		nextaddr, err := dasbuf.ReadString('\n')
		if err != nil {
			return err
		}
		relknode := libkiridir.PKeyLookup(nextaddr)
		if relknode == nil {
			return errors.New(fmt.Sprintf("Cannot find the relevant pubkey %s.", nextaddr))
		}
		// Establish connection to next node
		next_wire_raw, err := net.Dial("tcp", relknode.Address)
		if err != nil {
			return err
		}
		next_wire_actual, err := libkiss.Kiriobfs_handshake_client(next_wire_raw)
		if err != nil {
			return err
		}
		// Copy connections
		go func() {
			io.Copy(next_wire_actual, awire)
			next_wire_actual.Close()
		}()
		io.Copy(awire, next_wire_actual)
	} else {
		return errors.New("Unimplemented method in subcircuit requested")
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
