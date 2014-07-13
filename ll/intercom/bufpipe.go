package intercom

import (
	"io"
	"net"
	"sync"
	"time"
)

var _bplock sync.Mutex
var _bplist net.Listener

type BufferedPipe struct {
	buffer      []byte
	data_avail  chan bool
	lock        chan bool
	closed      bool
	buffer_free chan bool
}

func NewBufferedPipe() *BufferedPipe {
	toret := new(BufferedPipe)
	toret.data_avail = make(chan bool)
	toret.buffer = make([]byte, 0)
	toret.closed = false
	toret.lock = make(chan bool, 1)
	toret.lock <- true
	toret.buffer_free = make(chan bool, 1)
	return toret
}

func (pipe *BufferedPipe) Close() error {
	<-pipe.lock
	if pipe.closed {
		pipe.lock <- true
		return nil
	}
	pipe.closed = true
	pipe.lock <- true
	select {
	case pipe.data_avail <- false:
	default:
	}
	return nil
}

func (pipe *BufferedPipe) Write(p []byte) (int, error) {
	<-pipe.lock
	if pipe.closed {
		pipe.lock <- true
		return 0, io.ErrClosedPipe
	}
	if len(pipe.buffer) > 65536 {
		pipe.lock <- true
		<-pipe.buffer_free
		return pipe.Write(p)
	}

	pipe.buffer = append(pipe.buffer, p...)
	pipe.lock <- true
	select {
	case pipe.data_avail <- true:
	default:
	}
	return len(p), nil
}

func (pipe *BufferedPipe) Read(p []byte) (int, error) {
	<-pipe.lock
	if len(pipe.buffer) != 0 {
		cbts := copy(p, pipe.buffer)
		pipe.buffer = pipe.buffer[cbts:]
		pipe.lock <- true
		return cbts, nil
	}
	if pipe.closed {
		pipe.lock <- true
		return 0, io.EOF
	}
	//fmt.Println("telling people buff now free")
	select {
	case pipe.buffer_free <- true:
	default:
	}
	pipe.lock <- true
	select {
	case rslt := <-pipe.data_avail:
		if !rslt {
			return 0, io.EOF
		}
	case <-time.After(time.Second * 2):
	}
	return pipe.Read(p)
}
