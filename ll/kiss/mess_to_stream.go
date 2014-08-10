package kiss

import (
	"io"
	"kirisurf/ll/common"
)

// This file converts a strictly message-based ReadWriteCloser to a stream RWC.
// That is, it buffers reads.

// This interface is exported, since it is so useful.

type m2s_provider struct {
	underlying io.ReadWriteCloser
	buffer     *common.BufferedPipe
}

func MessToStream(mess io.ReadWriteCloser) io.ReadWriteCloser {
	var toret m2s_provider
	toret.underlying = mess
	toret.buffer = common.NewBufferedPipe()
	go func() {
		bts := make([]byte, 65536)
		defer toret.underlying.Close()
		defer toret.buffer.Close()
		for {
			n, err := toret.underlying.Read(bts)
			if err != nil {
				return
			}
			_, err = toret.buffer.Write(bts[:n])
			if err != nil {
				return
			}
		}
	}()
	return &toret
}

func (prov *m2s_provider) Read(p []byte) (int, error) {
	return prov.buffer.Read(p)
}

func (prov *m2s_provider) Write(p []byte) (int, error) {
	return prov.underlying.Write(p)
}

func (prov *m2s_provider) Close() error {
	return prov.underlying.Close()
}
