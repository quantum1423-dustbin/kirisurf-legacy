// server.go
package subcircuit

import (
	"fmt"
	"io"
	"libkiridir"
	"net"

	"code.google.com/p/log4go"
)

//TODO: Implement a timeout
func handleClient(client net.Conn) {
	log4go.Debug("Accepting subcircuit connection from %s", client.RemoteAddr())
	defer client.Close()
	protviol := func(e string) {
		log4go.Debug(e)
		fmt.Fprintf(client, "Protocol violation: %s\n", e)
	}
	command := make([]byte, 5)
	// read the four-byte command
	_, err := io.ReadFull(client, command)
	if err != nil {
		protviol(err.Error())
		return
	}
	cstr := string(command)
	if cstr == "ECHO\n" {
		log4go.Debug("%s requesting sc echo", client.RemoteAddr())
		io.Copy(client, client)
	} else if cstr == "CONN\n" {
		log4go.Debug("%s requesting sc remote connect...", client.RemoteAddr())
		var remoteaddr string
		_, err = fmt.Fscanln(client, &remoteaddr)
		if err != nil {
			protviol(err.Error())
			return
		}
		rconn, err := net.Dial("tcp", remoteaddr)
		if err != nil {
			return
		}
		log4go.Debug("%s requesting sc remote connect to %s", client.RemoteAddr(), remoteaddr)
		go func() {
			io.Copy(rconn, client)
			rconn.Close()
		}()
		io.Copy(client, rconn)
	} else if cstr == "KCON\n" {
		log4go.Debug("%s requesting sc kiri connect...", client.RemoteAddr())
		var remoteaddr string
		_, err = fmt.Fscanln(client, &remoteaddr)
		if err != nil {
			protviol(err.Error())
			return
		}
		log4go.Debug("%s requesting sc kiri connect to %s", client.RemoteAddr(), remoteaddr)
		rnode := libkiridir.PKeyLookup(remoteaddr)
		if rnode == nil {
			return
		}
		remoteaddr = rnode.Address
		log4go.Debug("%s requesting sc kiri connect to %s", client.RemoteAddr(), remoteaddr)
		rconn, err := net.Dial("tcp", remoteaddr)
		if err != nil {
			return
		}
		go func() {
			io.Copy(rconn, client)
			rconn.Close()
		}()
		io.Copy(client, rconn)
	} else {
		protviol("unrecognized command")
	}
}

func TestServer() {
	thing, err := net.Listen("tcp", "0.0.0.0:12345")
	if err != nil {
		log4go.Exit(err.Error())
	}
	for {
		c, e := thing.Accept()
		if e != nil {
			log4go.Error(e.Error())
			continue
		}
		handleClient(c)
	}
}
