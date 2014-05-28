package onionstew

import (
	"io"
	"math/rand"
	"time"

	"github.com/KirisurfProject/kilog"
)

import cryptorand "crypto/rand"

func (ctx *sc_ctx) AttachSC(wire io.ReadWriteCloser, serverside bool) {
	kilog.Debug("AttachSC(%v)", serverside)
	ctx.lock.Lock()
	ctx.refcount++
	ctx.lock.Unlock()
	defer func() {
		ctx.lock.Lock()
		ctx.refcount--
		ctx.lock.Unlock()
	}()
	local_stop := make(chan bool)  // Signal once for close, synchronous
	local_close := make(chan bool) // Close to remove this sc from the premises, cleanly
	ctx.close_ch_ch <- local_close
	// Read from the other side
	go func() {
		id := rand.Int()
		for {
			newpkt, err := read_sc_message(wire)
			if err != nil {
				kilog.Warning("AttachSC encountered unexpected error %s on %x while READING, DESTROYING STEW",
					err.Error(), id)
				ctx.destroy()
				wire.Close()
				return
			}
			// Check for the dead seqnum
			if newpkt.seqnum == 0xFFFFFFFFFFFFFFFF {
				kilog.Debug("Close message received from remote in AttachSC on %x, signalling...", id)
				if serverside {
					local_stop <- true
					kilog.Debug("Close signal successful, sending bakk and returning from %x.", id)
					clmsg := sc_message{0xFFFFFFFFFFFFFFFF, []byte("")}
					write_sc_message(clmsg, wire)
					time.Sleep(time.Second * 10)
					wire.Close()
				} else {
					wire.Close()
				}
				return
			}
			// Check for ignorable message
			if newpkt.seqnum == 0xFFFFFFFFFFFFFFFE {
				continue
			}
			select {
			case ctx.unordered_ch <- newpkt:
			case <-ctx.killswitch:
				kilog.Debug("Great, we got a KILLSWITCH instead of being able to put into unordered, fml")
				wire.Close()
				return
			}
		}
	}()
	// Write to the other side
	for {
		select {
		case newthing := <-ctx.write_ch:
			err := write_sc_message(newthing, wire)
			if err != nil {
				kilog.Warning("AttachSC encountered unexpected error %s while WRITING, DESTROYING STEW",
					err.Error())
				ctx.destroy()
				// Will die on next iteration
			}

			//TODO: Make this not tie up the connection
			if false {
				xaxa := make([]byte, 1)
				cryptorand.Read(xaxa)
				tosend := make([]byte, 0)
				for len(newthing.payload) < 1024 && xaxa[0] < 128 {
					tosend = append(tosend, make([]byte, xaxa[0])...)
				}
				qaqa := sc_message{0xFFFFFFFFFFFFFFFE, tosend}
				if len(tosend) != 0 {
					err := write_sc_message(qaqa, wire)
					if err != nil {
						kilog.Warning(
							"AttachSC encountered unexpected error %s while WRITING KA, DESTROYING STEW",
							err.Error())
						ctx.destroy()
					}
				}
			}
		case <-local_stop:
			return
		case <-local_close:
			kilog.Debug("AttachSC receiving LOCAL_CLOSE, stopping flow & sending remote")
			clmsg := sc_message{0xFFFFFFFFFFFFFFFF, []byte("")}
			write_sc_message(clmsg, wire)
			return
		case <-ctx.killswitch:
			kilog.Debug("AttachSC receiving KILLSWITCH, destroying wire")
			wire.Close()
			return
		case <-time.After(time.Second * time.Duration(rand.Int()%30)):
			xaxa := sc_message{0xFFFFFFFFFFFFFFFE, []byte("")}
			err := write_sc_message(xaxa, wire)
			if err != nil {
				kilog.Warning("AttachSC encountered unexpected error %s while WRITING KA, DESTROYING STEW",
					err.Error())
				ctx.destroy()
				// Will die on next iteration
			}
		}
	}
}
