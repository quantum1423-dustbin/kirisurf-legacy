// e2e_server.go
package main

import (
	"io"
	"kirisurf-legacy/ll/onionstew"
	"net"

	"github.com/KirisurfProject/kilog"
)

var serverport = ""

// e2e server handler. Subcircuit calls this.
func e2e_server_handler_old(wire io.ReadWriteCloser) {
	defer wire.Close()
	realconn, err := net.Dial("tcp", serverport)
	if err != nil {
		return
	}
	defer realconn.Close()
	go func() {
		defer wire.Close()
		io.Copy(wire, realconn)
	}()
	io.Copy(realconn, wire)
}

func init() {
	serverport = onionstew.RunManagedStewServer()
	kilog.Debug("After serverport")
}
