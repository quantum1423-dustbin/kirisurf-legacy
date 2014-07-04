package kiss

import (
	"crypto/rand"
	"crypto/subtle"
	"io"
	"net"
)

// This file implements obfs4f, which adds to obfs3f by adding an extension to
// the address format that makes sure any unauthorized probes are met with silence.

func Obfs4fHandshake(wire net.Conn, is_server bool, key string) (io.ReadWriteCloser, error) {
	secret := KeyedHash([]byte(key), []byte("Generator for obfs4f secret"))
	if is_server {
		// Must receive a valid proof!
		randdat := make([]byte, 256)
		cheksum := make([]byte, 64)
		_, err := io.ReadFull(wire, randdat)
		if err != nil {
			return nil, err
		}
		_, err = io.ReadFull(wire, cheksum)
		if err != nil {
			return nil, err
		}
		actcheksum := KeyedHash(randdat, secret)
		if subtle.ConstantTimeCompare(actcheksum, cheksum) == 1 {
			return Obfs3fHandshake(wire, is_server)
		} else {
			return nil, ErrMacNoMatch
		}
	} else {
		// Must send a valid proof!
		randdat := make([]byte, 256)
		rand.Read(randdat)
		cheksum := KeyedHash(randdat, secret)
		_, err := wire.Write(randdat)
		if err != nil {
			return nil, err
		}
		_, err = wire.Write(cheksum)
		if err != nil {
			return nil, err
		}
		return Obfs3fHandshake(wire, is_server)
	}
}
