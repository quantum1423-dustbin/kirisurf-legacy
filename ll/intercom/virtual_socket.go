package intercom

import "io"

type virtsock struct {
	reader *BufferedPipe
	writer *BufferedPipe
}

func (xaxa *virtsock) Read(p []byte) (int, error) {
	return xaxa.reader.Read(p)
}

func (xaxa *virtsock) Write(p []byte) (int, error) {
	return xaxa.writer.Write(p)
}

func (xaxa *virtsock) Close() error {
	xaxa.reader.Close()
	xaxa.writer.Close()
	return nil
}

func (xaxa *virtsock) Flipped() *virtsock {
	toret := new(virtsock)
	toret.reader = xaxa.writer
	toret.writer = xaxa.reader
	return toret
}

func new_vs() *virtsock {
	return &virtsock{NewBufferedPipe(), NewBufferedPipe()}
}

type VirtualServer struct {
	vs_ch chan *virtsock
}

func (vs *VirtualServer) Accept() (io.ReadWriteCloser, error) {
	toret := <-vs.vs_ch
	return toret, nil
}

func VSConnect(tgt *VirtualServer) io.ReadWriteCloser {
	toret := new_vs()
	tosend := toret.Flipped()
	tgt.vs_ch <- tosend
	return toret
}

func VSListen() *VirtualServer {
	return &VirtualServer{make(chan *virtsock)}
}
