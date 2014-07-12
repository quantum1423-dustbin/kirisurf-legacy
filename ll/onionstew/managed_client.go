package onionstew

import (
	"crypto/rand"
	"errors"
	"io"
)

// An onionstew.ManagedClient object represents an onion stew client to a single
// remote endpoint. Object can be destroyed, and clients can be funnelled in.
// All other operations are automated.

type ManagedClient struct {
	stew_id          []byte
	underlying       *stew_ctx
	client_chan      chan io.ReadWriteCloser
	client_addr_chan chan string
	DeadChan         chan bool
}

func MakeManagedClient(sc_generate func() io.ReadWriteCloser) (*ManagedClient, error) {

	toret := new(ManagedClient)
	toret.stew_id = make([]byte, 16)
	rand.Reader.Read(toret.stew_id)
	toret.underlying = make_stew_ctx()
	toret.client_chan = make(chan io.ReadWriteCloser)
	toret.client_addr_chan = make(chan string)
	toret.DeadChan = make(chan bool)

	go func() {
		<-toret.underlying.killswitch
		close(toret.DeadChan)
	}()

	go toret.underlying.run_stew(false)
	first_sc := sc_generate()
	_, err := first_sc.Write(toret.stew_id)
	if err != nil {
		return nil, errors.New("Cannot create SC successfully! (%s)")
	}
	go toret.underlying.llctx.AttachSC(first_sc, false)

	go func() {
		// 6 subcircuits
		for i := 0; i < 6; i++ {
			first_sc := sc_generate()
			_, err := first_sc.Write(toret.stew_id)
			if err != nil {
				toret.underlying.destroy()
				return
			}

			go toret.underlying.llctx.AttachSC(first_sc, false)
		}
	}()

	// Spin off a goroutine that accepts new clients
	go func() {
		for {
			client, ok := <-toret.client_chan
			if !ok {
				toret.underlying.destroy()
				return
			}
			remaddr := <-toret.client_addr_chan
			go toret.underlying.attacht_client(client, remaddr)
		}
	}()
	return toret, nil
}

func (thing *ManagedClient) Destroy() {
	close(thing.client_chan)
}

func (thing *ManagedClient) AddClient(ga io.ReadWriteCloser, remaddr string) {
	thing.client_chan <- ga
	thing.client_addr_chan <- remaddr
}
