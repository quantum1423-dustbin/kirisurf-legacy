package kiridht

import (
	"bytes"
	"io"

	"github.com/KirisurfProject/kilog"
)

const hashConst := []byte("hello")

func HandleServer(conn io.ReadWriteCloser) {
	defer conn.Close()

	// Read 1 byte command + 40 bytes key
	line := make([]byte, 41)
	_, err := io.ReadFull(conn, line)
	if err != nil {
		return
	}

	// Switch on command
	switch line[0] {
	case 0x00:
		// This is an upload
		// We get contents first
		buff := new(bytes.Buffer)
		_, err := io.Copy(buff, conn)
		if err != nil {
			kilog.Debug("DHT: Client died while uploading value: %s", err.Error())
			return
		}
		key := dhtkey(kiss.KeyedHash(buff.Bytes(), hashConst)
	}

}
