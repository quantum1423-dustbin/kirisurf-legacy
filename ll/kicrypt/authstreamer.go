// authstreamer.go
package kicrypt

import (
	"crypto/cipher"
	"crypto/rc4"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
)

// Authenticated streamer. Counter nonce is implicit!
type AuthStreamer interface {
	Open(ct []byte) ([]byte, error)
	Seal(pt []byte) ([]byte, error)
}

// AEAD-based streamer
type __AEAD_authstreamer struct {
	counter   uint64
	aeadstate cipher.AEAD
}

// XOR stream cipher struct with HMAC-SHA512
type __XOR_authstreamer struct {
	counter     uint64
	secret_key  []byte
	streamstate cipher.Stream
}

func (state __XOR_authstreamer) Open(ct []byte) ([]byte, error) {
	defer func() {
		state.counter++
	}()
	counterbytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterbytes, state.counter)
	// of authenticatings
	actual_hash := hash_keyed(ct[64:], string(append(counterbytes, state.secret_key...)))
	purported_hash := ct[:64]
	if subtle.ConstantTimeCompare(actual_hash, purported_hash) != 1 {
		return nil, errors.New("HMAC mismatch!")
	}
	state.streamstate.XORKeyStream(ct[64:], ct[64:])
	return ct[64:], nil
}

func (state __XOR_authstreamer) Seal(pt []byte) ([]byte, error) {
	defer func() {
		state.counter++
	}()
	toret := make([]byte, 64+len(pt))
	// of encrypt
	state.streamstate.XORKeyStream(toret[64:], pt)
	// of authenticate
	counterbytes := make([]byte, 8)
	noncebytes := string(append(counterbytes, state.secret_key...))
	hash := hash_keyed(toret[64:], string(noncebytes))
	copy(toret[0:64], hash)
	return toret, nil
}

func (state __AEAD_authstreamer) Open(ct []byte) ([]byte, error) {
	defer func() {
		state.counter++
	}()
	nonce := make([]byte, state.aeadstate.NonceSize())
	binary.LittleEndian.PutUint64(nonce, state.counter)
	return state.aeadstate.Open(make([]byte, 0), nonce, ct, make([]byte, 0))
}

func (state __AEAD_authstreamer) Seal(pt []byte) ([]byte, error) {
	defer func() {
		state.counter++
	}()
	nonce := make([]byte, state.aeadstate.NonceSize())
	binary.LittleEndian.PutUint64(nonce, state.counter)
	return state.aeadstate.Seal(make([]byte, 0), nonce, pt, make([]byte, 0)), nil
}

func AS_arcfour128_drop8192(key []byte) AuthStreamer {
	cryptkey := hash_keyed(key, "cryptkey")
	hashkey := hash_keyed(key, "hashkey")
	streamstate, _ := rc4.NewCipher(cryptkey[:16])
	truestate := cipher.Stream(streamstate)
	junk := make([]byte, 8192)
	truestate.XORKeyStream(junk, junk)
	toret := __XOR_authstreamer{
		0,
		hashkey,
		truestate}
	return AuthStreamer(toret)
}

func AS_aes128_gcm(key []byte) AuthStreamer {
	cryptkey := hash_keyed(key, "cryptkey")[:16]
	gcmstate := fastAES_GCM(cryptkey)
	toret := __AEAD_authstreamer{0, gcmstate}
	return AuthStreamer(toret)
}

func AS_aes128_ctr(key []byte) AuthStreamer {
	cryptkey := hash_keyed(key, "cryptkey")[:16]
	hashkey := hash_keyed(key, "hashkey")
	blockstate := fastAES_initialize(cryptkey)
	streamstate := cipher.NewCTR(blockstate, make([]byte, blockstate.BlockSize()))
	truestate := cipher.Stream(streamstate)
	toret := __XOR_authstreamer{
		0,
		hashkey,
		truestate}
	return AuthStreamer(toret)
}

func AS_blowfish128_ofb(key []byte) AuthStreamer {
	cryptkey := hash_keyed(key, "cryptkey")[:16]
	hashkey := hash_keyed(key, "hashkey")
	streamstate := fastBF_NewOFB(cryptkey)
	truestate := cipher.Stream(streamstate)
	toret := __XOR_authstreamer{
		0,
		hashkey,
		truestate}
	return AuthStreamer(toret)
}
