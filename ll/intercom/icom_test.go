package intercom

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"
)

func TestICOM_HelloWorld(t *testing.T) {
	listener := MakeIntercomServer("booger@127.0.0.1:12346")
	go func() {
		for {
			thing := listener.Accept()
			rand.Seed(time.Now().UnixNano())
			randnum := rand.Int() % 100
			fmt.Fprintf(thing, " *** The following sequence should end at %d ***\n", randnum)
			for i := 0; i <= randnum; i++ {
				fmt.Fprintf(thing, "%d ", i)
			}
			fmt.Fprintf(thing, "\n")
			thing.Close()
		}
	}()
	dialer := MakeIntercomDialer()
	connection, err := dialer.Dial("booger@127.0.0.1:12346")
	if err != nil {
		fmt.Println(err.Error())
		t.FailNow()
	}
	io.Copy(os.Stdout, connection)
}

func TestICOM_13371(t *testing.T) {
	t.Skip("This test is interactive!")
	listener := MakeIntercomServer("stuff@127.0.0.1:1234")
	go func() {
		for {
			thing := listener.Accept()
			go func() {
				defer thing.Close()
				remote, err := net.Dial("tcp", "127.0.0.1:13370")
				if err != nil {
					panic(err.Error())
				}
				defer remote.Close()
				go func() {
					defer thing.Close()
					io.Copy(thing, remote)
				}()
				io.Copy(remote, thing)
			}()
		}
	}()

	clist, err := net.Listen("tcp", "127.0.0.1:13371")
	dialer := MakeIntercomDialer()
	if err != nil {
		panic(err.Error())
	}
	for {
		thing, err := clist.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer thing.Close()
			remote, err := dialer.Dial("stuff@127.0.0.1:1234")
			if err != nil {
				panic(err.Error())
			}
			defer remote.Close()
			go func() {
				defer thing.Close()
				io.Copy(thing, remote)
			}()
			io.Copy(remote, thing)
		}()
	}
}
