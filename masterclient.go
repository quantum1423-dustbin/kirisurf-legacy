// masterclient.go
package main

import (
	"kirisurf/ll/dirclient"
	"math/rand"
	"net"
	"runtime"

	"github.com/coreos/go-log/log"
)

var ctx_buffer = make(chan e2e_client_ctx, 9)

func enfreshen_scb() {
	log.Alert("Enfreshen!")
	for i := 0; i < 5; i++ {
		go func() {
		retry:
			log.Alert("Enfreshen!")
			thing, err := build_subcircuit()
			if err != nil {
				dirclient.RefreshDirectory()
				goto retry
			}
			ctx_buffer <- make_e2e_client_ctx(thing.wire)
		}()
	}
	log.Alert("Freshened!")
}

func run_client_loop() {
	enfreshen_scb()
	// Round robin, basically
	var get_ctx func() e2e_client_ctx
	get_ctx = func() e2e_client_ctx {
		toret := <-ctx_buffer
		if !*toret.valid || *toret.dying {
			log.Debug("BUFFERED CTX NOT VALID")
			go func() {
			retry:
				log.Alert("Enfreshen!")
				thing, err := build_subcircuit()
				if err != nil {
					goto retry
				}
				ctx_buffer <- make_e2e_client_ctx(thing.wire)
			}()
			return get_ctx()
		}
		ctx_buffer <- toret
		if rand.Int()%20 == 0 {
			log.Debug("MARKING AS DYING")
			*toret.dying = true
			return get_ctx()
		}
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

func run_diagnostic_loop() {
	listener, err := net.Listen("tcp", "127.0.0.1:9222")
	if err != nil {
		panic(err)
	}
	for {
		nconn, err := listener.Accept()
		if err != nil {
			log.Debug("Error while accepting: ", err.Error())
			continue
		}
		go func() {
			defer nconn.Close()
			buf := make([]byte, 65536)
			n := runtime.Stack(buf, true)
			nconn.Write(buf[:n])
		}()
	}
}
