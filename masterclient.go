// masterclient.go
package main

import (
	"fmt"
	"io"
	"kirisurf/ll/circuitry"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/intercom"
	"kirisurf/ll/socks5"
	"net"
	"sync"
	"time"

	"github.com/KirisurfProject/kilog"
)

var circ_ch chan intercom.MultiplexClient

func produce_circ() intercom.MultiplexClient {
	xaxa := dirclient.FindExitPath(MasterConfig.Network.MinCircuitLen)
	lel, err := circuitry.BuildCircuit(xaxa, 254)
	if err != nil {
		dirclient.RefreshDirectory()
		time.Sleep(time.Second)
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
	kilog.Info("Bootstrapping 100%%: client started!")

	go func() {
		var haha sync.WaitGroup
		haha.Add(5)
		for i := 0; i < 5; i++ {
			go func() {
				circ_ch <- produce_circ()
				haha.Done()
			}()
		}
		haha.Wait()
	}()
	for {
		nconn, err := listener.Accept()
		if err != nil {
			kilog.Warning("Problem while accepting client socket: %s", err.Error())
			continue
		}
		addr, err := socks5.ReadRequest(nconn)
		if err != nil {
			kilog.Warning("Problem while reading SOCKS5 request")
			nconn.Close()
			continue
		}
		go func() {
			defer func() {
				nconn.Close()
			}()
		retry:
			newcirc := <-circ_ch
			remote, err := newcirc.SocksAccept(nconn)
			if err != nil {
				dirclient.RefreshDirectory()
				circ_ch <- produce_circ()
				goto retry
			}
			circ_ch <- newcirc
			defer remote.Close()
			lenbts := []byte{byte((len(addr) + 1) % 256), byte((len(addr) + 1) / 256)}
			_, err = remote.Write(lenbts)
			_, err = remote.Write([]byte(fmt.Sprintf("t%s", addr)))
			if err != nil {
				kilog.Debug("Failed to send tunnelling request to %s!", addr)
				socks5.CompleteRequest(0x03, nconn)
				return
			}

			code := make([]byte, 4)
			_, err = io.ReadFull(remote, code)
			if err != nil {
				kilog.Debug("Failed to read response for %s!", addr)
				socks5.CompleteRequest(0x03, nconn)
				return
			}

			switch string(code) {
			case "OKAY":
				kilog.Debug("Successfully tunneled %s!", addr)
				socks5.CompleteRequest(0x00, nconn)
				go func() {
					defer remote.Close()
					io.Copy(remote, nconn)
				}()
				io.Copy(nconn, remote)
			case "TMOT":
				kilog.Debug("Tunnel to %s timed out!", addr)
				socks5.CompleteRequest(0x06, nconn)
			case "NOIM":
				kilog.Debug("Tunnel type for %s isn't implemented by server!", addr)
				socks5.CompleteRequest(0x07, nconn)
			case "FAIL":
				kilog.Debug("Tunnel to %s cannot be established!", addr)
				socks5.CompleteRequest(0x04, nconn)
			default:
				kilog.Debug("Protocol error on tunnel to %s!", addr)
				socks5.CompleteRequest(0x01, nconn)
				return
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
			kilog.Warning("Problem while accepting stacktrace diag socket: %s", err.Error())
			continue
		}
		go func() {
			defer nconn.Close()
			for {
				str := <-kilog.FineChannel
				_, err := nconn.Write([]byte(fmt.Sprintf("%s\n", str)))
				if err != nil {
					return
				}
			}
		}()
	}
}

func init() {
	circ_ch = make(chan intercom.MultiplexClient, 8)
}
