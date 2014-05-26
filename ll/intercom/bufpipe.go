package intercom

import (
	"io"
	"sync"
	"time"
)

type BufferedPipe struct {
	fail_ch  chan bool
	main_ch  chan []byte
	buffer   []byte
	read_lk  sync.Mutex
	write_lk sync.Mutex
	xaxa     sync.Once
}

func NewBufferedPipe() *BufferedPipe {
	toret := new(BufferedPipe)
	toret.buffer = nil
	toret.fail_ch = make(chan bool)
	toret.main_ch = make(chan []byte, 4)
	return toret
}

func (pipe *BufferedPipe) Close() error {
	pipe.xaxa.Do(func() {
		time.Sleep(time.Second / 20)
		close(pipe.fail_ch)
	})
	return nil
}

func (pipe *BufferedPipe) Write(p []byte) (int, error) {
	pipe.write_lk.Lock()
	defer pipe.write_lk.Unlock()

	select {
	case pipe.main_ch <- p:
		return len(p), nil
	case <-pipe.fail_ch:
		return 0, io.ErrClosedPipe
	}
}

func (pipe *BufferedPipe) Read(p []byte) (int, error) {
	pipe.read_lk.Lock()
	defer pipe.read_lk.Unlock()

	// If something in buffer...
retry:
	if pipe.buffer != nil {
		if len(pipe.buffer) > len(p) {
			copy(p, pipe.buffer)
			pipe.buffer = pipe.buffer[len(p):]
			return len(p), nil
		}
		copy(p, pipe.buffer)
		return len(pipe.buffer), nil
	}

	var arrival []byte
	select {
	case arrival = <-pipe.main_ch:
	case <-pipe.fail_ch:
		return 0, io.EOF
	}
	if len(arrival) > len(p) {
		if pipe.buffer == nil {
			pipe.buffer = arrival
			goto retry
		} else {
			panic("Kurwa! This shouldn't into happen!")
		}
	}
	copy(p, arrival)
	return len(arrival), nil
}
