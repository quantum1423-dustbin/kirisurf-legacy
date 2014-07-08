package intercom

import (
	"io"
	"net"
	"sync"
)

var _bplock sync.Mutex
var _bplist net.Listener

type BufferedPipe struct {
	reader io.ReadWriteCloser
	writer io.ReadWriteCloser
	wmutex sync.Mutex
	rmutex sync.Mutex
}

func NewBufferedPipe() *BufferedPipe {
	_bplock.Lock()
	defer _bplock.Unlock()
	writer, err := net.Dial("tcp", _bplist.Addr().String())
	if err != nil {
		panic("wtf")
	}
	reader, err := _bplist.Accept()
	if err != nil {
		panic("wtf")
	}
	toret := new(BufferedPipe)
	toret.reader = reader
	toret.writer = writer
	return toret
}

func (pipe *BufferedPipe) Read(p []byte) (int, error) {
	//pipe.rmutex.Lock()
	//defer pipe.rmutex.Unlock()
	return pipe.reader.Read(p)
}

func (pipe *BufferedPipe) Write(p []byte) (int, error) {
	//pipe.wmutex.Lock()
	//defer pipe.wmutex.Unlock()
	return pipe.writer.Write(p)
}

func (pipe *BufferedPipe) Close() error {
	//pipe.wmutex.Lock()
	//defer pipe.wmutex.Unlock()
	return pipe.writer.Close()
}

func init() {
	_bplist, _ = net.Listen("tcp", "127.0.0.1:0")
}
