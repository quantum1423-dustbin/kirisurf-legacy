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
	/*
		go func() {
			for i := 0; i < 5; i++ {
				circ_ch <- produce_circ()
			}
		}()*/
	for {
		nconn, err := listener.Accept()
		if err != nil {
			WARNING("Problem while accepting client socket: %s", err.Error())
			continue
		}
		go func() {
			defer nconn.Close()
			newcirc := <-circ_ch
			remote, err := newcirc.SocksAccept(nconn)
			if err != nil {
				dirclient.RefreshDirectory()
				go func() {
					circ_ch <- produce_circ()
				}()
			}
			circ_ch <- newcirc
			defer remote.Close()
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
