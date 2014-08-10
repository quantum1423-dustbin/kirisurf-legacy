package common

import (
	"io"
	"net"
	"sync"
)

var _bplock sync.Mutex
var _bplist net.Listener

type BufferedPipe struct {
	buffer      []byte
	data_avail  chan bool
	lock        chan bool
	closed      bool
	close_ch    chan bool
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
	toret.close_ch = make(chan bool)
	return toret
}

func (pipe *BufferedPipe) Close() error {
	<-pipe.lock
	if pipe.closed {
		pipe.lock <- true
		return nil
	}
	close(pipe.close_ch)
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
		select {
		case <-pipe.buffer_free:
		case <-pipe.close_ch:
			return 0, io.ErrClosedPipe
		}
		return pipe.Write(p)
	}

	oldlen := len(pipe.buffer)
	pipe.buffer = append(pipe.buffer, p...)
	pipe.lock <- true
	if oldlen == 0 {
		pipe.data_avail <- true
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
	case <-pipe.close_ch:
		return 0, io.EOF
	}
	return pipe.Read(p)
}
