// masterclient.go
package main

import (
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/onionstew"
	"math/rand"
	"net"
	"runtime"
	"time"

	"github.com/KirisurfProject/kilog"
)

var theBigContext *onionstew.ManagedClient
var viableNodes [][]dirclient.KNode

func enfreshen_scb() {
	// We shouldn't enfreshen unless the existing ctx is dead
	if theBigContext != nil {
		<-theBigContext.DeadChan
	}

	// Refresh the directory & viable nodes
	dirclient.RefreshDirectory()
	viableNodes = dirclient.FindPathGroup(MasterConfig.Network.MinCircuitLen)

	// Function for returnings of one stronk subcircuit
	gen_subcircuit := func() io.ReadWriteCloser {
	retry:
		xaxa := viableNodes[rand.Int()%16]
		sc, err := build_subcircuit(xaxa)
		if err != nil {
			kilog.Warning("What? %v", err.Error())
			goto retry
		}
		return sc
	}

	tbc, err := onionstew.MakeManagedClient(gen_subcircuit)
	if err != nil {
		kilog.Warning("error encountered in enfreshen_scb() %s, sleeping 1 sec & retry", err.Error())
		time.Sleep(time.Second)
		enfreshen_scb()
		return
	}
	theBigContext = tbc
}

func run_client_loop() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	pxlistener, err := net.Listen("tcp", MasterConfig.General.SocksAddr)
	if err != nil {
		panic(err)
	}
	enfreshen_scb()
	go func() {
		for {
			enfreshen_scb()
		}
	}()
	set_gui_progress(1.0)
	INFO("Bootstrapping 100%%: client started!")
	go func() {
		for {
			nconn, err := listener.Accept()
			if err != nil {
				WARNING("Problem while accepting client socket: %s", err.Error())
				continue
			}
			go func() {
				remaddr, err := socks5_handshake(nconn)
				if err != nil {
					nconn.Close()
					return
				}
				kilog.Debug("remaddr=%s", remaddr)
				theBigContext.AddClient(nconn, remaddr)
			}()
		}
	}()
	for {
		nconn, err := pxlistener.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer nconn.Close()
			act, err := net.Dial("tcp", listener.Addr().String())
			if err != nil {
				return
			}
			defer act.Close()
			go func() {
				defer nconn.Close()
				defer act.Close()
				buf := make([]byte, 8192)
				for {
					n, err := nconn.Read(buf)
					if err != nil {
						return
					}
					incr_up_bytes(n)
					_, err = act.Write(buf[:n])
					if err != nil {
						return
					}
				}
			}()
			buf := make([]byte, 8192)
			for {
				n, err := act.Read(buf)
				if err != nil {
					return
				}
				incr_down_bytes(n)
				_, err = nconn.Write(buf[:n])
				if err != nil {
					return
				}
			}
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
