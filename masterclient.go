// masterclient.go
package main

import (
	"fmt"
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/onionstew"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/KirisurfProject/kilog"
)

var theBigContext *onionstew.ManagedClient
var viableNodes [][]dirclient.KNode

var tbclock sync.Mutex

func enfreshen_scb() {
	// We shouldn't enfreshen unless the existing ctx is dead
	if theBigContext != nil {
		kilog.Debug("Waiting for dead chan...")
		<-theBigContext.DeadChan
	}
	time.Sleep(time.Second * 2)
	tbclock.Lock()
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
			time.Sleep(time.Second)
			dirclient.RefreshDirectory()
			viableNodes = dirclient.FindPathGroup(MasterConfig.Network.MinCircuitLen)
			goto retry
		}
		return sc
	}

	fmt.Println("Got to 1")
	tbc, err := onionstew.MakeManagedClient(gen_subcircuit)
	fmt.Println("Got to 2")
	if err != nil {
		kilog.Warning("error encountered in enfreshen_scb() %s, sleeping 1 sec & retry", err.Error())
		time.Sleep(time.Second)
		enfreshen_scb()
		return
	}
	theBigContext = tbc
	fmt.Println("Got to 3")
	tbclock.Unlock()
}

func run_client_loop() {
	listener, err := net.Listen("tcp", MasterConfig.General.SocksAddr)
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
			kilog.Debug("Connecting to %s", remaddr)
			tbclock.Lock()
			tbc := theBigContext
			tbclock.Unlock()
			tbc.AddClient(nconn, remaddr)
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
