package onionstew

import (
	"encoding/binary"
	"io"
	"sync"
)

const (
	m_open  = 0x01
	m_close = 0x02
	m_data  = 0x03
	m_more  = 0x04
	m_dns   = 0x10
)

type stew_message struct {
	category int
	connid   int
	payload  []byte
}

func (thing stew_message) bytes() []byte {
	toret := make([]byte, len(thing.payload)+3)
	toret[0] = byte(thing.category)
	toret[1] = byte(thing.connid / 256)
	toret[2] = byte(thing.connid % 256)
	copy(toret[3:], thing.payload)
	return toret
}

func bytes_to_stew_message(thing []byte) stew_message {
	var toret stew_message
	toret.category = int(thing[0])
	toret.connid = int(thing[1])*256 + int(thing[2])
	toret.payload = thing[3:]
	return toret
}

type sc_message struct {
	seqnum  uint64
	payload []byte
}

func read_sc_message(src io.Reader) (sc_message, error) {
	var toret sc_message
	scratch := make([]byte, 65536)
	_, err := io.ReadFull(src, scratch[:8])
	if err != nil {
		return toret, err
	}
	seqnum := binary.BigEndian.Uint64(scratch[:8])
	_, err = io.ReadFull(src, scratch[:2])
	if err != nil {
		return toret, err
	}
	paylen := binary.BigEndian.Uint16(scratch[:2])
	_, err = io.ReadFull(src, scratch[:paylen])
	if err != nil {
		return toret, err
	}
	toret.seqnum = seqnum
	toret.payload = scratch[:paylen]
	return toret, nil
}

func write_sc_message(tow sc_message, dst io.Writer) error {
	headerlen := 8 + 2
	scratch := make([]byte, headerlen)
	binary.BigEndian.PutUint64(scratch[:8], tow.seqnum)
	binary.BigEndian.PutUint16(scratch[8:10], uint16(len(tow.payload)))
	_, err := dst.Write(scratch)
	if err != nil {
		return err
	}
	_, err = dst.Write(tow.payload)
	return err
}

type sc_ctx struct {
	lock         sync.Mutex
	refcount     int
	unordered_ch chan sc_message
	ordered_ch   chan sc_message
	write_ch     chan sc_message
	killswitch   chan bool
	destroy      func()
	close_ch_ch  chan chan bool
}

func make_sc_ctx() *sc_ctx {
	toret := new(sc_ctx)
	toret.unordered_ch = make(chan sc_message)
	toret.ordered_ch = make(chan sc_message)
	toret.killswitch = make(chan bool)
	toret.write_ch = make(chan sc_message)
	toret.close_ch_ch = make(chan chan bool, 256)
	go reorder_messages(toret.unordered_ch, toret.ordered_ch)
	go func() {
		<-toret.killswitch
		close(toret.unordered_ch)
	}()
	var thing sync.Once
	toret.destroy = func() {
		thing.Do(func() {
			close(toret.killswitch)
		})
	}
	return toret
}

type stew_ctx struct {
	lock       sync.RWMutex
	llctx      *sc_ctx
	conntable  [65536]chan stew_message
	killswitch chan bool
	write_ch   chan stew_message
	client_ch  chan io.ReadWriteCloser
	destroy    func()
	number_ch  chan int
}

func make_stew_ctx() *stew_ctx {
	toret := new(stew_ctx)
	toret.llctx = make_sc_ctx()
	toret.client_ch = make(chan io.ReadWriteCloser)
	toret.write_ch = make(chan stew_message)
	toret.number_ch = make(chan int, 65536)
	var xaxa sync.Once
	toret.destroy = func() {
		xaxa.Do(func() {
			toret.llctx.destroy()
			close(toret.killswitch)
		})
	}
	toret.killswitch = make(chan bool)
	// Kill when underlying dies
	go func() {
		<-toret.llctx.killswitch
		toret.destroy()
	}()
	go func() {
		ctr := uint64(0)
		for {
			select {
			case <-toret.killswitch:
				return
			case thing := <-toret.write_ch:
				//kilog.Debug("Writing stew_message{%d, %d, %s} with ctr %d", thing.category,
				//	thing.connid, string(thing.payload), ctr)
				scm := sc_message{ctr, thing.bytes()}
				select {
				case <-toret.killswitch:
					return
				case toret.llctx.write_ch <- scm:
					//	kilog.Debug("Write with ctr=%d went through!", ctr)
				}
			}
			ctr++
		}
	}()

	for i := 0; i < 16384; i++ {
		toret.number_ch <- i
	}
	return toret
}
