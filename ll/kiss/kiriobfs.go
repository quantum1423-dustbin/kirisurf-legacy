// kiriobfs
package kiss

import (
	"crypto/rand"
	"crypto/rc4"
	"errors"
	"io"
	"kirisurf/ll/kicrypt"
	"math/big"
	"net"
	"time"
)

// Kirisurf_handshake_server returns a socket obfuscating the socket sock after
// initiating the client-side handshake.
func Kiriobfs_handshake_server(sock net.Conn) (net.Conn, error) {
	// generate our keypair
	our_keypair := kicrypt.UniformDH_genpair()
	scatch := make(chan bool, 16)

	// read their key
	buf := make([]byte, 192)
	_, ecd := io.ReadFull(sock, buf)
	if ecd != nil {
		LOG(LOG_ERROR, "protocol mismatch for obfuscation")
		return nil, ecd
	}
	LOG(LOG_DEBUG, "Read their %d-byte key", 192)
	theirpub := big.NewInt(0).SetBytes(buf)

	//send our key
	c, err := sock.Write(our_keypair.PublicBytes())
	LOG(LOG_DEBUG, "Sent our %d-byte key", c)
	if err != nil {
		return nil, errors.New("error encountered when sending uniformdh pubkey")
	}

	//send 16384 bytes of garbage
	go func() {
		for i := 0; i < 64; i++ {
			buf := make([]byte, 256)
			rand.Reader.Read(buf)
			c, err := sock.Write(buf)
			if err != nil && c == c {
				return
			}
		}
		scatch <- true
	}()

	//read 16384 bytes of garbage from remote
	go func() {
		buf := make([]byte, 16384)
		_, err := io.ReadFull(sock, buf)
		if err != nil {
			return
		}
		scatch <- true
	}()

	//synchronize
	<-scatch
	<-scatch
	LOG(LOG_DEBUG, "All garbage-related things done.")

	//shared secret
	shared_secret := kicrypt.UniformDH_gensecret(our_keypair.Private, theirpub)
	up_key := kicrypt.KeyedHash(shared_secret, "kiriobfs/up")
	down_key := kicrypt.KeyedHash(shared_secret, "kiriobfs/down")
	LOG(LOG_DEBUG, "Shared secret derived.")

	//RC4 initialization
	var toret Kiriobfs_state
	toret.RC4_state_r, err = rc4.NewCipher(up_key)
	toret.RC4_state_w, err = rc4.NewCipher(down_key)
	toret.underlying = sock
	throwaway := make([]byte, 8192)
	toret.RC4_state_r.XORKeyStream(throwaway, throwaway)
	toret.RC4_state_w.XORKeyStream(throwaway, throwaway)

	//Actual garbage
	blarg := make([]byte, 2)
	rand.Reader.Read(blarg)
	bllen := big.NewInt(0).SetBytes(blarg).Int64()
	toret.Write(blarg)
	toret.Write(make([]byte, bllen))
	LOG(LOG_DEBUG, "Encrypted junk sent.")

	//Read the actual garbage
	glarg := make([]byte, 2)
	io.ReadFull(net.Conn(toret), glarg)
	glwerg := make([]byte, big.NewInt(0).SetBytes(glarg).Int64())
	io.ReadFull(net.Conn(toret), glwerg)
	return net.Conn(toret), nil
}

// Kirisurf_handshake_client returns a socket obfuscating the socket sock after
// initiating the client-side handshake.
func Kiriobfs_handshake_client(sock net.Conn) (net.Conn, error) {
	// generate our keypair
	our_keypair := kicrypt.UniformDH_genpair()
	scatch := make(chan bool, 16)

	//send our key
	c, err := sock.Write(our_keypair.PublicBytes())
	LOG(LOG_DEBUG, "Sent our %d-byte key", c)
	if err != nil {
		return nil, errors.New("error encountered when sending uniformdh pubkey")
	}

	//send 16384 bytes of garbage
	go func() {
		for i := 0; i < 64; i++ {
			buf := make([]byte, 256)
			rand.Reader.Read(buf)
			c, err := sock.Write(buf)
			if err != nil && c == c {
				return
			}
		}
		scatch <- true
	}()

	// read their key
	buf := make([]byte, 192)
	_, ecd := io.ReadFull(sock, buf)
	if ecd != nil {
		return nil, ecd
	}
	LOG(LOG_DEBUG, "Read their %d-byte key", 192)
	theirpub := big.NewInt(0).SetBytes(buf)

	//read 16384 bytes of garbage from remote
	go func() {
		buf := make([]byte, 16384)
		_, err := io.ReadFull(sock, buf)
		if err != nil {
			return
		}
		scatch <- true
	}()

	//synchronize
	<-scatch
	<-scatch
	LOG(LOG_DEBUG, "All garbage-related things done.")

	//shared secret
	shared_secret := kicrypt.UniformDH_gensecret(our_keypair.Private, theirpub)
	up_key := kicrypt.KeyedHash(shared_secret, "kiriobfs/up")
	down_key := kicrypt.KeyedHash(shared_secret, "kiriobfs/down")
	LOG(LOG_DEBUG, "Shared secret derived.")

	//RC4 initialization
	var toret Kiriobfs_state
	toret.RC4_state_r, err = rc4.NewCipher(down_key)
	toret.RC4_state_w, err = rc4.NewCipher(up_key)
	toret.underlying = sock
	throwaway := make([]byte, 8192)
	toret.RC4_state_r.XORKeyStream(throwaway, throwaway)
	toret.RC4_state_w.XORKeyStream(throwaway, throwaway)

	//Actual garbage
	blarg := make([]byte, 2)
	rand.Reader.Read(blarg)
	bllen := big.NewInt(0).SetBytes(blarg).Int64()
	toret.Write(blarg)
	toret.Write(make([]byte, bllen))
	LOG(LOG_DEBUG, "Encrypted junk sent.")

	//Read the actual garbage
	glarg := make([]byte, 2)
	io.ReadFull(net.Conn(toret), glarg)
	glwerg := make([]byte, big.NewInt(0).SetBytes(glarg).Int64())
	io.ReadFull(net.Conn(toret), glwerg)
	return net.Conn(toret), nil
}

/* The following defines the Kiriobfs_state structure. */

type Kiriobfs_state struct {
	RC4_state_r *rc4.Cipher
	RC4_state_w *rc4.Cipher
	underlying  net.Conn
}

func (state Kiriobfs_state) Read(b []byte) (int, error) {
	n, err := state.underlying.Read(b)
	//state.RC4_state_r.Reset()
	state.RC4_state_r.XORKeyStream(b[:n], b[:n])
	return n, err
}

func (state Kiriobfs_state) Write(b []byte) (int, error) {
	buf := make([]byte, len(b))
	//state.RC4_state_w.Reset()
	state.RC4_state_w.XORKeyStream(buf, b)
	return state.underlying.Write(buf)
}

func (state Kiriobfs_state) Close() error {
	return state.underlying.Close()
}

func (state Kiriobfs_state) LocalAddr() net.Addr {
	return state.underlying.LocalAddr()
}

func (state Kiriobfs_state) RemoteAddr() net.Addr {
	return state.underlying.RemoteAddr()
}

func (state Kiriobfs_state) SetDeadline(t time.Time) error {
	return state.underlying.SetDeadline(t)
}

func (state Kiriobfs_state) SetReadDeadline(t time.Time) error {
	return state.underlying.SetReadDeadline(t)
}

func (state Kiriobfs_state) SetWriteDeadline(t time.Time) error {
	return state.underlying.SetWriteDeadline(t)
}
