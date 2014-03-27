// masterclient.go
package main

import (
	"net"

	"github.com/coreos/go-log/log"
)

var ctx_buffer = make(chan e2e_client_ctx, 9)

func enfreshen_scb() {
	log.Alert("Enfreshen!")
	for i := 0; i < 3; i++ {
	retry:
		log.Alert("Enfreshen!")
		thing, err := build_subcircuit()
		if err != nil {
			goto retry
		}
		ctx_buffer <- make_e2e_client_ctx(thing.wire)
	}
	log.Alert("Freshened!")
}

func run_client_loop() {
	enfreshen_scb()
	// Round robin, basically
	get_ctx := func() e2e_client_ctx {
		toret := <-ctx_buffer
		ctx_buffer <- toret
		return toret
	}
	// Main loop
	listener, err := net.Listen("tcp", MasterConfig.General.SocksAddr)
	if err != nil {
		panic(err)
	}
	for {
		nconn, err := listener.Accept()
		log.Debug("Accepted client with address: ", nconn.RemoteAddr())
		if err != nil {
			log.Debug("Error while accepting: ", err.Error())
			continue
		}
		go get_ctx().AttachClient(nconn)
	}
}
