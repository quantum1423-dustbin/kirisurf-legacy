// e2e_client.go
package main

import (
	"io"
	"sync"
)

type e2e_client_ctx struct {
	connid_chan chan int
	chan_table  map[int]chan e2e_segment
	wire        *gobwire
	lock        *sync.RWMutex
	valid       *bool
}

func make_e2e_client_ctx(conn io.ReadWriteCloser) e2e_client_ctx {
	wire := newGobWire(conn)
	lock := new(sync.RWMutex)
	valid := true
	chan_table := make(map[int]chan e2e_segment)
	connid_chan := make(chan int, 65536)
	for i := 0; i < 65536; i++ {
		connid_chan <- i
	}
	return e2e_client_ctx{connid_chan, chan_table, wire, lock, &valid}
}

func (ctx e2e_client_ctx) AttachClient(client io.ReadWriteCloser) {
	if !*ctx.valid {
		panic("Context already invalid!")
	}
	// Obtain a connection ID
	connid := <-ctx.connid_chan
	ch := make(chan e2e_segment, 1024)
	// Attach onto channel table
	ctx.lock.Lock()
	ctx.chan_table[connid] = ch
	ctx.lock.Unlock()
	// Detach function
	detach := func() {
		ctx.lock.Lock()
		ctx.chan_table[connid] = nil
		ctx.lock.Unlock()
	}
	// Downstream
	go func() {
		defer client.Close()
		defer detach()
		for {
			if !*ctx.valid {
				ctx.wire.destroy()
				return
			}
			pkt, ok := <-ch
			if !ok {
				ctx.wire.destroy()
				*ctx.valid = false
				return
			}
			if pkt.flag == E2E_CLOSE {
				return
			}
			_, err := client.Write(pkt.body)
			if err != nil {
				return
			}
		}
	}()
	defer client.Close()
	defer detach()
	// Upstream
	for {
		if !*ctx.valid {
			return
		}
		buf := make([]byte, 16384)
		n, err := client.Read(buf)
		if err != nil {
			err := ctx.wire.Send(e2e_segment{E2E_CLOSE, connid, []byte("")})
			if err != nil {
				*ctx.valid = false
				ctx.wire.destroy()
			}
			return
		}
		err = ctx.wire.Send(e2e_segment{E2E_DATA, connid, buf[:n]})
		if err != nil {
			*ctx.valid = false
			ctx.wire.destroy()
			return
		}
	}
}
