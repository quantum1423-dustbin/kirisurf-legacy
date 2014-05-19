package kiss

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/rc4"
	"io"
)

// This file implements obfs3f, a lighter and less fingerprinty derivative of Tor's obfs3.
// RC4/drop8192 replaces AES-CTR, and the handshake sees the server remaining silent until client sends full.
// f is for fast. Designed to be fast enough for transparent encapsulation of everything.

// This is a public interface.

type Obfs3f struct {
	read_rc4   cipher.Stream
	write_rc4  cipher.Stream
	underlying io.ReadWriteCloser
}

func Obfs3fHandshake(wire io.ReadWriteCloser, is_server bool) (io.ReadWriteCloser, error) {
	var their_public dh_public_key
	var our_keypair = dh_gen_key(1536)
	var secret []byte
	var write_rc4, read_rc4 cipher.Stream
	their_public = make([]byte, 1536/8)
	if is_server {
		_, err := io.ReadFull(wire, their_public)
		if err != nil {
			return nil, err
		}
		_, err = wire.Write(our_keypair.Public)
		if err != nil {
			return nil, err
		}
		secret = dh_gen_secret(our_keypair.Private, their_public)
		write_rc4, _ = rc4.NewCipher(KeyedHash(secret, []byte("obfs3f_downstr")))
		read_rc4, _ = rc4.NewCipher(KeyedHash(secret, []byte("obfs3f_upstr")))
	} else {
		_, err := wire.Write(our_keypair.Public)
		if err != nil {
			return nil, err
		}
		_, err = io.ReadFull(wire, their_public)
		if err != nil {
			return nil, err
		}
		secret = dh_gen_secret(our_keypair.Private, their_public)
		write_rc4, _ = rc4.NewCipher(KeyedHash(secret, []byte("obfs3f_upstr")))
		read_rc4, _ = rc4.NewCipher(KeyedHash(secret, []byte("obfs3f_downstr")))
	}
	read_rc4.XORKeyStream(make([]byte, 8192), make([]byte, 8192))
	write_rc4.XORKeyStream(make([]byte, 8192), make([]byte, 8192))

	toret := &Obfs3f{read_rc4, write_rc4, wire}
	thing := make(chan bool)
	go func() {
		randlen := make([]byte, 2)
		rand.Read(randlen)
		rlint := int(randlen[0])*256 + int(randlen[1])
		xaxa := make([]byte, rlint)
		toret.Write(randlen)
		toret.Write(xaxa)
		thing <- true
	}()

	randlen := make([]byte, 2)
	_, err := io.ReadFull(toret, randlen)
	if err != nil {
		return nil, err
	}
	_, err = io.ReadFull(toret, make([]byte, int(randlen[0])*256+int(randlen[1])))
	if err != nil {
		return nil, err
	}
	<-thing
	return io.ReadWriteCloser(toret), nil
}

func (thing *Obfs3f) Write(p []byte) (int, error) {
	xaxa := make([]byte, len(p))
	copy(xaxa, p)
	thing.write_rc4.XORKeyStream(xaxa, xaxa)
	return thing.underlying.Write(xaxa)
}

func (thing *Obfs3f) Read(p []byte) (int, error) {
	n, err := thing.underlying.Read(p)
	if err != nil {
		return 0, err
	}
	thing.read_rc4.XORKeyStream(p[:n], p[:n])
	return n, nil
}

func (thing *Obfs3f) Close() error {
	return thing.underlying.Close()
}
