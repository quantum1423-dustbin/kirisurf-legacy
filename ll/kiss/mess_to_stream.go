package kiss

import (
	"fmt"
	"io"
	"sync"
)

// This file converts a strictly message-based ReadWriteCloser to a stream RWC.
// That is, it buffers reads.

// This interface is exported, since it is so useful.

type m2s_provider struct {
	underlying io.ReadWriteCloser
	buffer     []byte
	rlock      sync.Mutex
	wlock      sync.Mutex
}

func MessToStream(mess io.ReadWriteCloser) io.ReadWriteCloser {
	var toret m2s_provider
	toret.underlying = mess
	return &toret
}

func (prov *m2s_provider) Read(p []byte) (int, error) {
	prov.rlock.Lock()
	defer prov.rlock.Unlock()
	if prov.buffer != nil {
		copy(p, prov.buffer)
		if len(p) < len(prov.buffer) {
			prov.buffer = prov.buffer[len(p):]
			return len(p), nil
		} else {
			prov.buffer = nil
			return len(prov.buffer), nil
		}
	}
	rbuf := make([]byte, 65536)
	n, err := prov.underlying.Read(rbuf)
	if err != nil {
		return 0, err
	}
	if n <= len(p) {
		copy(p, rbuf)
		return n, nil
	}
	can_be_read := rbuf[:len(p)]
	must_be_buf := rbuf[len(p):n]
	copy(p, can_be_read)
	prov.buffer = must_be_buf
	fmt.Println("Put some in buffer")
	return len(can_be_read), nil
}

func (prov *m2s_provider) Write(p []byte) (int, error) {
	return prov.underlying.Write(p)
}

func (prov *m2s_provider) Close() error {
	return prov.underlying.Close()
}
