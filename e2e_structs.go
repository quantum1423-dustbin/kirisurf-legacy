// e2e_structs.go
package main

import (
	"encoding/binary"
	"io"
	"sync"

	"github.com/coreos/go-log/log"
)

// Structures for the end-to-end protocol.

const (
	__INVALID = iota
	E2E_OPEN  = iota
	E2E_CLOSE = iota
	E2E_DATA  = iota
	E2E_ECHO  = iota
)

type e2e_segment struct {
	Flag   int
	Connid int
	Body   []byte
}

type gobwire struct {
	conn    io.ReadWriteCloser
	_rlock  *sync.Mutex
	_slock  *sync.Mutex
	destroy func() error
}

func (wire *gobwire) Receive() (e2e_segment, error) {
	wire._rlock.Lock()
	defer wire._rlock.Unlock()
	var placeholder e2e_segment
	b_flag := make([]byte, 1)
	b_connid := make([]byte, 2)
	b_length := make([]byte, 2)
	_, err := io.ReadFull(wire.conn, b_flag)
	if err != nil {
		log.Debug("Receive() returning an error: ", err.Error())
		return placeholder, err
	}
	_, err = io.ReadFull(wire.conn, b_connid)
	if err != nil {
		log.Debug("Receive() returning an error: ", err.Error())
		return placeholder, err
	}
	_, err = io.ReadFull(wire.conn, b_length)
	if err != nil {
		log.Debug("Receive() returning an error: ", err.Error())
		return placeholder, err
	}
	b_body := make([]byte, binary.LittleEndian.Uint16(b_length))
	_, err = io.ReadFull(wire.conn, b_body)
	if err != nil {
		log.Debug("Receive() returning an error: ", err.Error())
		return placeholder, err
	}
	return e2e_segment{int(b_flag[0]), int(binary.LittleEndian.Uint16(b_connid)), b_body}, nil
}

func (wire *gobwire) Send(thing e2e_segment) error {
	wire._slock.Lock()
	defer wire._slock.Unlock()
	//log.Debug("Sending: ", thing)
	tosend := make([]byte, len(thing.Body)+1+2+2)
	tosend[0] = byte(thing.Flag)
	binary.LittleEndian.PutUint16(tosend[1:3], uint16(thing.Connid))
	binary.LittleEndian.PutUint16(tosend[3:5], uint16(len(thing.Body)))
	copy(tosend[5:], thing.Body)
	_, err := wire.conn.Write(tosend)
	//log.Debug("Sent.")
	return err
}

func newGobWire(thing io.ReadWriteCloser) *gobwire {
	toret := gobwire{thing, new(sync.Mutex), new(sync.Mutex), thing.Close}
	return &toret
}
