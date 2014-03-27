// e2e_structs.go
package main

import (
	"encoding/gob"
	"io"
	"sync"
)

// Structures for the end-to-end protocol.

const (
	E2E_OPEN  = iota
	E2E_CLOSE = iota
	E2E_DATA  = iota
)

type e2e_segment struct {
	Flag   int
	Connid int
	Body   []byte
}

type gobwire struct {
	in      *gob.Decoder
	out     *gob.Encoder
	_lock   *sync.Mutex
	destroy func() error
}

func (wire *gobwire) Receive() (e2e_segment, error) {
	wire._lock.Lock()
	defer wire._lock.Unlock()
	var toret e2e_segment
	err := wire.in.Decode(&toret)
	if err != nil {
		return toret, err
	}
	return toret, nil
}

func (wire *gobwire) Send(thing e2e_segment) error {
	wire._lock.Lock()
	defer wire._lock.Unlock()
	return wire.out.Encode(thing)
}

func newGobWire(thing io.ReadWriteCloser) *gobwire {
	toret := gobwire{gob.NewDecoder(thing), gob.NewEncoder(thing), new(sync.Mutex), thing.Close}
	return &toret
}
