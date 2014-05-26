package kiss

import (
	"errors"
	"fmt"
	"io"
)

// This file implements the transport of kiss. Handshake is completely symmetrical; nonces ensure
// no ugly things involving keys on both sides happen.

type kiss_mess_ctx struct {
	read_crypter  *chugger
	write_crypter *chugger
	underlying    io.ReadWriteCloser
}

// KiSS handshake

func TransportHandshake(keypair DHKeys, wire io.ReadWriteCloser,
	verify func([]byte) bool) (io.ReadWriteCloser, error) {

	eph_keypair := dh_gen_key(2048)
	done := make(chan bool)
	go func() {
		// Send longterm public key
		_, err := wire.Write(keypair.Public)
		if err != nil {
			return
		}
		// Send ephemeral public key
		_, err = wire.Write(eph_keypair.Public)
		if err != nil {
			return
		}
		done <- true
	}()
	// Read longterm public key
	their_pubkey := make([]byte, 2048/8)
	_, err := io.ReadFull(wire, their_pubkey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Couldn't read their longterm key (%s)", err.Error()))
	}
	if !verify(their_pubkey) {
		return nil, ErrMacNoMatch
	}

	// Read ephemeral public key
	their_eph_pubkey := make([]byte, 2048/8)
	_, err = io.ReadFull(wire, their_eph_pubkey)
	if err != nil {
		return nil, errors.New("Couldn't read their shortterm key")
	}

	// Generate secret
	secret1 := dh_gen_secret(keypair.Private, their_pubkey)
	secret2 := dh_gen_secret(eph_keypair.Private, their_eph_pubkey)
	secret := append(secret1, secret2...)

	// Generate r and w secrets by appending THEIR eph pubkey for w, and OUR eph pubkey for r
	write_secret := append(secret, their_eph_pubkey...)
	read_secret := append(secret, eph_keypair.Public...)

	write_key := KeyedHash(write_secret, []byte("KiSS-1.0"))
	read_key := KeyedHash(read_secret, []byte("KiSS-1.0"))

	readk := new([32]byte)
	copy(readk[:], read_key)
	writek := new([32]byte)
	copy(writek[:], write_key)
	<-done
	return MessToStream(&kiss_mess_ctx{&chugger{readk, 0, 0},
		&chugger{writek, 0, 0}, wire}), nil
}

// KiSS transport is simply 2 bytes BE length + payload.

func (ctx *kiss_mess_ctx) Read(p []byte) (int, error) {
	lenthing := make([]byte, 2)
	_, err := io.ReadFull(ctx.underlying, lenthing)
	if err != nil {
		return 0, err
	}
	actlen := int(int(lenthing[0])*int(256)) + int(lenthing[1])
	payload := make([]byte, actlen)
	_, err = io.ReadFull(ctx.underlying, payload)
	if err != nil {
		return 0, err
	}
	toret, err := ctx.read_crypter.Open(payload)
	if err != nil {
		return 0, err
	}
	if len(toret) > len(p) {
		panic("READ BUFFER TOO SMALL KURWA KURWA KURWA")
	}
	copy(p, toret)
	return len(toret), nil
}

func (ctx *kiss_mess_ctx) Write(p []byte) (int, error) {
	lenthing := make([]byte, 2)
	crypted := ctx.write_crypter.Seal(p)
	lenthing[0], lenthing[1] = byte(len(crypted)/256), byte(len(crypted)%256)
	_, err := ctx.underlying.Write(lenthing)
	if err != nil {
		return 0, err
	}
	_, err = ctx.underlying.Write(crypted)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (ctx *kiss_mess_ctx) Close() error {
	return ctx.underlying.Close()
}
