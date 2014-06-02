package kiss

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func run_single_kiss_echo(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	defer listener.Close()
	client, err := listener.Accept()
	if err != nil {
		panic(err.Error())
	}
	defer client.Close()
	xaxa, err := Obfs3fHandshake(client, true)
	if err != nil {
		panic(err.Error())
	}
	xaxa, err = TransportHandshake(dh_gen_key(2048), xaxa,
		func([]byte) bool { return true })
	xaxa, err = TransportHandshake(dh_gen_key(2048), xaxa,
		func([]byte) bool { return true })
	xaxa, err = TransportHandshake(dh_gen_key(2048), xaxa,
		func([]byte) bool { return true })
	xaxa, err = TransportHandshake(dh_gen_key(2048), xaxa,
		func([]byte) bool { return true })
	if err != nil {
		panic(err.Error())
	}
	defer xaxa.Close()
	ggg := make([]byte, 8192)
	for {
		_, err := xaxa.Write(ggg)
		if err != nil {
			return
		}
	}
}

func BenchmarkKiSS(b *testing.B) {
	go run_single_kiss_echo("127.0.0.1:1234")
	time.Sleep(time.Second)
	lololol, err := net.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		panic("wtf")
	}
	defer lololol.Close()
	xaxa, err := Obfs3fHandshake(lololol, false)
	if err != nil {
		panic(err.Error())
	}
	xaxa, err = TransportHandshake(dh_gen_key(2048), xaxa,
		func([]byte) bool { return true })
	xaxa, err = TransportHandshake(dh_gen_key(2048), xaxa,
		func([]byte) bool { return true })
	xaxa, err = TransportHandshake(dh_gen_key(2048), xaxa,
		func([]byte) bool { return true })
	xaxa, err = TransportHandshake(dh_gen_key(2048), xaxa,
		func([]byte) bool { return true })
	if err != nil {
		panic(err.Error())
	}
	defer xaxa.Close()
	fmt.Printf("xaxa!\n")
	listener, _ := net.Listen("tcp", "127.0.0.1:5555")
	client, _ := listener.Accept()
	defer client.Close()
	go io.Copy(client, xaxa)
	io.Copy(xaxa, client)
}
