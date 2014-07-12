// masterclient.go
package main

import (
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/intercom"
	"net"
	"runtime"
)

var circ_ch chan intercom.MultiplexClient

func produce_circ() intercom.MultiplexClient {
	xaxa := dirclient.FindOnePath(MasterConfig.Network.MinCircuitLen)
	lel, err := build_subcircuit(xaxa)
	if err != nil {
		dirclient.RefreshDirectory()
		return produce_circ()
	}
	return intercom.MakeMultiplexClient(lel)
}

func run_client_loop() {
	listener, err := net.Listen("tcp", MasterConfig.General.SocksAddr)
	if err != nil {
		panic(err)
	}
	circ_ch <- produce_circ()
	set_gui_progress(1.0)
	INFO("Bootstrapping 100%%: client started!")
	for {
		nconn, err := listener.Accept()
		if err != nil {
			WARNING("Problem while accepting client socket: %s", err.Error())
			continue
		}
		go func() {
			defer nconn.Close()
			newcirc := <-circ_ch
			circ_ch <- newcirc
			remote, err := newcirc.SocksAccept(nconn)
			defer remote.Close()
			if err != nil {
				panic("Can only panic for now!")
			}
			go func() {
				defer remote.Close()
				io.Copy(remote, nconn)
			}()
			io.Copy(nconn, remote)
		}()
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

func init() {
	circ_ch = make(chan intercom.MultiplexClient, 8)
}
