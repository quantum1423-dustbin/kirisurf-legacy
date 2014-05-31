// clientcomm.go
package main

import (
	"fmt"
	"net"
	"sync"

	"kirisurf/ll/onionstew"
)

var global_monitor_chan = make(chan []byte, 16)

var global_down_bytes = 0
var global_up_bytes = 0

var global_locker sync.Mutex

func set_gui_progress(frac float64) {
	msg := []byte(fmt.Sprintf("(set-progress %f)\n", frac))
	select {
	case global_monitor_chan <- msg:
	default:
	}
}

func incr_down_bytes(delta int) {
	global_locker.Lock()
	global_down_bytes += delta
	global_locker.Unlock()
	msg := []byte(fmt.Sprintf("(incr-download %d)\n", delta))
	select {
	case global_monitor_chan <- msg:
	default:
	}
}

func incr_down_overhead_bytes(delta int) {
	global_locker.Lock()
	global_down_bytes += delta
	global_locker.Unlock()
	msg := []byte(fmt.Sprintf("(incr-download-overhead %d)\n", delta))
	select {
	case global_monitor_chan <- msg:
	default:
	}
}
func incr_up_overhead_bytes(delta int) {
	global_locker.Lock()
	global_down_bytes += delta
	global_locker.Unlock()
	msg := []byte(fmt.Sprintf("(incr-upload-overhead %d)\n", delta))
	select {
	case global_monitor_chan <- msg:
	default:
	}
}

func incr_up_bytes(delta int) {
	global_locker.Lock()
	global_up_bytes += delta
	global_locker.Unlock()
	msg := []byte(fmt.Sprintf("(incr-upload %d)\n", delta))
	select {
	case global_monitor_chan <- msg:
	default:
	}
}

func run_monitor_loop() {
	onionstew.DownloadIncrement = incr_down_bytes
	onionstew.DownloadOverheadIncrement = incr_down_overhead_bytes
	onionstew.UploadIncrement = incr_up_bytes
	onionstew.UploadOverheadIncrement = incr_up_overhead_bytes

	listener, err := net.Listen("tcp", "127.0.0.1:9221")
	if err != nil {
		panic(err.Error())
	}
	for {
		client, err := listener.Accept()
		if err != nil {
			WARNING(err.Error())
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
