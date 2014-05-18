package onionstew

import (
	"crypto/rand"
	"errors"
	"io"
	"time"

	"github.com/KirisurfProject/kilog"
)

// An onionstew.ManagedClient object represents an onion stew client to a single
// remote endpoint. Object can be destroyed, and clients can be funnelled in.
// All other operations are automated.

type ManagedClient struct {
	stew_id     []byte
	underlying  *stew_ctx
	client_chan chan io.ReadWriteCloser
	DeadChan    chan bool
}

func MakeManagedClient(sc_generate func() io.ReadWriteCloser) (*ManagedClient, error) {
	rand256 := func() int {
		gah := make([]byte, 1)
		rand.Read(gah)
		return int(gah[0])
	}

	toret := new(ManagedClient)
	toret.stew_id = make([]byte, 16)
	rand.Reader.Read(toret.stew_id)
	toret.underlying = make_stew_ctx()
	toret.client_chan = make(chan io.ReadWriteCloser)
	toret.DeadChan = toret.underlying.killswitch
	go toret.underlying.run_stew(false)

	// First subcircuit
	first_sc := sc_generate()
	_, err := first_sc.Write(toret.stew_id)
	if err != nil {
		return nil, errors.New("Cannot create SC successfully! (%s)")
	}

	go toret.underlying.llctx.AttachSC(first_sc, false)

	// Spin off a goroutine that constantly adds/removes subcircuits
	go func() {
		ctr := 1
		for {
			// Sleep between 1 and 64 seconds
			select {
			case <-time.After(time.Second * 4):
			case <-toret.underlying.killswitch:
				return
			}
			octr := ctr
			if ctr < 3 {
				// Too little subcircuits, add one!
				sc := sc_generate()
				_, err := sc.Write(toret.stew_id)
				if err != nil {
					kilog.Warning("What? Subcircuit add failed!")
				}
				go toret.underlying.llctx.AttachSC(sc, false)
				ctr++
			} else if ctr > 5 {
				// Too many subcircuits, remove one!
				close(<-toret.underlying.llctx.close_ch_ch)
				ctr--
			} else {
				// Randomly decide
				if rand256() < 128 {
					sc := sc_generate()
					_, err := sc.Write(toret.stew_id)
					if err != nil {
						kilog.Warning("What? Subcircuit add failed!")
					}
					go toret.underlying.llctx.AttachSC(sc, false)
					ctr++
				} else {
					close(<-toret.underlying.llctx.close_ch_ch)
					ctr--
				}
			}
			kilog.Debug("Subcircuit count changed from %d to %d!", octr, ctr)
		}
	}()

	// Spin off a goroutine that accepts new clients
	go func() {
		for {
			client, ok := <-toret.client_chan
			defer client.Close()
			if !ok {
				toret.underlying.destroy()
				return
			}
			go toret.underlying.attacht_client(client)
		}
	}()
	return toret, nil
}

func (thing *ManagedClient) Destroy() {
	close(thing.client_chan)
}

func (thing *ManagedClient) AddClient(ga io.ReadWriteCloser) {
	thing.client_chan <- ga
}
