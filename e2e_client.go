// e2e_client.go
package main

import (
	"io"
	"sync"
	"sync/atomic"
)

type e2e_client_ctx struct {
	connid_chan chan int
	chan_table  map[int]chan e2e_segment
	wire        *gobwire
	lock        *sync.RWMutex
	valid       *bool
	dying       *bool
	refcount    *int32
}

func make_e2e_client_ctx(conn io.ReadWriteCloser) e2e_client_ctx {
	wire := newGobWire(conn)
	lock := new(sync.RWMutex)
	valid := new(bool)
	*valid = true
	chan_table := make(map[int]chan e2e_segment)
	connid_chan := make(chan int, 65536)
	for i := 0; i < 65536; i++ {
		connid_chan <- i
	}
	// Loop that pushes data onto clients
	go func() {
		for {
			if !*valid {
				return
			}
			newpkt, err := wire.Receive()
			if err != nil {
				*valid = false
				wire.destroy()
				return
			}
			if chan_table[newpkt.Connid] == nil {
				continue
			}
			chan_table[newpkt.Connid] <- newpkt
		}
	}()
	dying := new(bool)
	*dying = false
	refcount := new(int32)
	*refcount = 0
	return e2e_client_ctx{connid_chan, chan_table, wire, lock, valid, dying, refcount}
}

func (ctx e2e_client_ctx) AttachClient(client io.ReadWriteCloser) {
	defer client.Close()
	if !*ctx.valid {
		CRITICAL("Possible race condition: AttachClient called on invalid wire!")
		return
	}
	atomic.AddInt32(ctx.refcount, 1)
	defer func() {
		atomic.AddInt32(ctx.refcount, -1)
		if *ctx.refcount == 0 && *ctx.dying {
			DEBUG("Killing a subcircuit context due to refcount")
			*ctx.valid = false
			ctx.lock.RLock()
			for _, e := range ctx.chan_table {
				if e != nil {
					close(e)
				}
			}
			ctx.lock.RUnlock()
			ctx.wire.destroy()
		}
	}()

	// SOCKS5 stuff! Yay!
	remaddr, err := socks5_handshake(client)
	if err != nil {
		WARNING("Error encountered while doing socks5: %s", err.Error())
		return
	}
	DEBUG("SOCKS5 request to %s", remaddr)

	defer func() {
		DEBUG("Closed connection to %s", remaddr)
	}()

	// Obtain a connection ID
	connid := <-ctx.connid_chan
	ch := make(chan e2e_segment, 256)
	ctr := 0
	// Attach onto channel table
	ctx.lock.Lock()
	ctx.chan_table[connid] = ch
	ctx.lock.Unlock()
	// Detach function
	var once sync.Once
	detach := func() {
		once.Do(func() {
			ctx.lock.Lock()
			ctx.chan_table[connid] = nil
			close(ch)
			ctx.lock.Unlock()
		})
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
				return
			}
			if pkt.Flag == E2E_CLOSE {
				return
			}
			n, err := client.Write(pkt.Body)
			if err != nil {
				return
			}
			incr_down_bytes(n)
			ctr = (ctr + 1) % 256
			// If wire of empty, sendings of sendmore
			if ctr == 0 {
				DEBUG("Bucket drained in subcircuit, sending SENDMORE to remote")
				err = ctx.wire.Send(e2e_segment{E2E_SENDMORE, connid, []byte("")})
				if err != nil {
					panic(err.Error())
				}
			}
		}
	}()
	defer detach()
	// Upstream
	err = ctx.wire.Send(e2e_segment{E2E_OPEN, connid, []byte(remaddr)})
	if err != nil {
		CRITICAL(err.Error())
		return
	}
	for {
		if !*ctx.valid {
			DEBUG("Dying since ctx not valid!!!")
			return
		}
		buf := make([]byte, 16384)
		n, err := client.Read(buf)
		if err != nil {
			err := ctx.wire.Send(e2e_segment{E2E_CLOSE, connid, []byte("")})
			if err != nil {
				DEBUG("Dying since cannot into sendings: %s", err.Error())
				*ctx.valid = false
				ctx.wire.destroy()
			}
			return
		}
		err = ctx.wire.Send(e2e_segment{E2E_DATA, connid, buf[:n]})
		if err != nil {
			DEBUG("Dying since cannot into sendings: %s", err.Error())
			*ctx.valid = false
			ctx.wire.destroy()
			return
		}
		incr_up_bytes(n)
	}
}
