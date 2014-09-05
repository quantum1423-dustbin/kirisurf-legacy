package kiridht

import (
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	thing, _ := net.Listen("tcp", "127.0.0.1:1234")
	conn, _ := thing.Accept()
	HandleServer(conn)
}
