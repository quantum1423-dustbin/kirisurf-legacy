package intercom

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestVS(t *testing.T) {
	server := VSListen()
	go func() {
		xaxa, _ := server.Accept()
		defer xaxa.Close()
		for i := 0; i < 100; i++ {
			xaxa.Write([]byte(fmt.Sprintf("Hello world! %d\n", i)))
		}
	}()
	conn := VSConnect(server)
	_, err := io.Copy(os.Stdout, conn)
	if err != nil {
		panic(err.Error())
	}
}
