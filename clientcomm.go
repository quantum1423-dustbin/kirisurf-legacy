// clientcomm.go
package main

import (
	"fmt"
	"net"

	"github.com/coreos/go-log/log"
)

var global_monitor_chan = make(chan []byte, 16)

func set_gui_progress(frac float64) {
	msg := []byte(fmt.Sprintf("(set-progress %f)\n", frac))
	select {
	case global_monitor_chan <- msg:
	default:
	}
}

func incr_down_bytes(delta int) {
	msg := []byte(fmt.Sprintf("(incr-download %d)\n", delta))
	select {
	case global_monitor_chan <- msg:
	default:
	}
}

func incr_up_bytes(delta int) {
	msg := []byte(fmt.Sprintf("(incr-upload %d)\n", delta))
	select {
	case global_monitor_chan <- msg:
	default:
	}
}

func run_monitor_loop() {
	listener, err := net.Listen("tcp", "127.0.0.1:9221")
	if err != nil {
		panic(err.Error())
	}
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Error(err.Error())
			continue
		}
		func() {
			defer client.Close()
			for {
				thing := <-global_monitor_chan
				_, err := client.Write(thing)
				if err != nil {
					return
				}
			}
		}()
	}
}
