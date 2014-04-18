// masterclient.go
package main

import (
	"kirisurf/ll/dirclient"
	"math/rand"
	"net"
	"runtime"
	"sync"
)

var ctx_buffer = make(chan e2e_client_ctx, 9)

func enfreshen_scb() {
	var wg sync.WaitGroup
	wg.Add(7)
	ctr := 0.3
	for i := 0; i < 6; i++ {
		i := i
		go func() {
			INFO("Building initial subcircuit #%d...", i)
		retry:
			thing, err := build_subcircuit()
			if err != nil {
				WARNING("Building of initial subcircuit #%d encountered trouble (%s), retrying...", i, err.Error())
				dirclient.RefreshDirectory()
				goto retry
			}
			ctx_buffer <- make_e2e_client_ctx(thing.wire)
			ctr = ctr + 0.1
			INFO("Bootstrapping %d%%: initial subcircuit %d done!", int(ctr*100), i)
			set_gui_progress(ctr)
			wg.Done()
		}()
	}
	wg.Wait()
}

func run_client_loop() {
	enfreshen_scb()
	set_gui_progress(1)
	INFO("Bootstrapping 100%%: inital circuits built, ready to go!")
	// Round robin, basically
	var get_ctx func() e2e_client_ctx
	get_ctx = func() e2e_client_ctx {
		toret := <-ctx_buffer
		if !*toret.valid || *toret.dying {
			DEBUG("Encountered dead subcircuit in buffer, throwing away")
			go func() {
			retry:
				thing, err := build_subcircuit()
				if err != nil {
					WARNING("Building of non-initial subcircuit encountered trouble, retrying...")
					dirclient.RefreshDirectory()
					goto retry
				}
				DEBUG("Queuing a new live subcircuit to buffer...")
				ctx_buffer <- make_e2e_client_ctx(thing.wire)
			}()
			return get_ctx()
		}
		ctx_buffer <- toret
		if rand.Int()%50 == 0 {
			DEBUG("Subcircuit lottery hit, marking as dead...")
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
		if err != nil {
			WARNING("Problem accepting client socket: %s", err.Error())
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
			WARNING("Problem while accepting stacktrace diag socket: %s", err.Error())
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
