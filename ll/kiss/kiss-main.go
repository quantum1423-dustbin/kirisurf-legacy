// kiss-main.go
package kiss

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"kirisurf/ll/kicrypt"
	"math/big"
)

type KiSS_State struct {
	read_ciph       kicrypt.AuthStreamer
	write_ciph      kicrypt.AuthStreamer
	shared_secret   []byte
	written_packets uint64
	read_packets    uint64
	wire            io.ReadWriteCloser
	buffer          *[]byte
}

type Verifier func(*big.Int) bool

func dumb_Verifier(*big.Int) bool {
	return true
}

var __KISS_AS func([]byte) kicrypt.AuthStreamer

func SetCipher(blah func([]byte) kicrypt.AuthStreamer) {
	__KISS_AS = blah
}

func KiSS_handshake_server(wire io.ReadWriteCloser, keys kicrypt.SecureDH_keypair) (io.ReadWriteCloser, error) {
	// read their handshake
	their_greeting_packed, err1 := KiSS_read_segment(wire)
	if err1 != nil {
		return nil, err1
	}
	if their_greeting_packed.segment_type != K_HANDSHAKE_C {
		return nil, errors.New("client sent something other than a handshake")
	}
	// unpack their greeting
	their_greeting, err2 := KiSS_unpack_client_handshake(their_greeting_packed.raw_payload)
	if err2 != nil {
		return nil, err2
	}
	// send our handshake
	our_keypair := kicrypt.SecureDH_genpair()
	our_greeting := KiSS_Segment{K_HANDSHAKE_S, (KiSS_HS_Server{keys.Public,
		our_keypair.Public}).Pack()}
	_, err3 := wire.Write(our_greeting.Bytes())
	if err3 != nil {
		return nil, err3
	}
	// obtain the shared secret
	secret := kicrypt.SecureDH_gensecret(our_keypair.Private, their_greeting.public_edh_key)
	secret = append(secret, kicrypt.SecureDH_gensecret(keys.Private, their_greeting.public_edh_key)...)
	upkey := kicrypt.KeyedHash(secret, "kiss-2-up")
	downkey := kicrypt.KeyedHash(secret, "kiss-2-down")
	read_ciph := __KISS_AS(upkey)
	write_ciph := __KISS_AS(downkey)
	bt := new([]byte)
	*bt = make([]byte, 0)
	toret := KiSS_State{read_ciph, write_ciph, secret, 0, 0, wire, bt}
	//LOG(LOG_DEBUG, "keygens on server side of done")
	return io.ReadWriteCloser(toret), nil
}

func KiSS_handshake_client(wire io.ReadWriteCloser, verify Verifier) (io.ReadWriteCloser, error) {
	// our keys
	our_keypair := kicrypt.SecureDH_genpair()
	//LOG(LOG_DEBUG, "CPUB: %d", our_keypair.Public)
	// construct the client handshake
	our_greeting := KiSS_HS_Client{0x02, our_keypair.Public}
	our_greeting_packed := KiSS_Segment{K_HANDSHAKE_C, our_greeting.Pack()}
	// send across the client handshake
	_, err := wire.Write(our_greeting_packed.Bytes())
	if err != nil {
		return nil, errors.New("wtf")
	}
	//LOG(LOG_DEBUG, "our client greeting of sent")
	// obtain the server handshake
	their_handshake_raw, err2 := KiSS_read_segment(wire)
	//LOG(LOG_DEBUG, "their server greeting of got")
	if err2 != nil {
		return nil, err2
	}
	if their_handshake_raw.segment_type != K_HANDSHAKE_S {
		return nil, errors.New("wtf")
	}
	their_handshake, err3 := KiSS_unpack_server_handshake(their_handshake_raw.raw_payload, verify)
	if err3 != nil {
		return nil, err3
	}
	// obtain the shared secret
	secret := kicrypt.SecureDH_gensecret(our_keypair.Private, their_handshake.public_edh_key)
	secret = append(secret, kicrypt.SecureDH_gensecret(our_keypair.Private, their_handshake.public_dh_key)...)
	upkey := kicrypt.KeyedHash(secret, "kiss-2-up")
	downkey := kicrypt.KeyedHash(secret, "kiss-2-down")
	// get the cipher thingies
	read_ciph := __KISS_AS(downkey)
	write_ciph := __KISS_AS(upkey)
	bt := new([]byte)
	*bt = make([]byte, 0)
	toret := KiSS_State{read_ciph, write_ciph, secret, 0, 0, wire, bt}
	return io.ReadWriteCloser(toret), nil
}

func (state KiSS_State) Write(p []byte) (int, error) {
	//LOG(LOG_DEBUG, "Entering KiSS_State.Write")
	defer func() {
		state.written_packets++
		//LOG(LOG_DEBUG, "Exiting KiSS_State.Write")
	}()
	towrite, _ := state.write_ciph.Seal(p)
	encaps := KiSS_Segment{K_APP_DATA, towrite}
	//LOG(LOG_DEBUG, "Written segment: %s|%X", encaps.StringRep(), encaps.Bytes())
	_, err := state.wire.Write(encaps.Bytes())
	return len(p), err
}

func (state KiSS_State) Close() error {
	//LOG(LOG_DEBUG, "Closing a KiSS connection")
	*state.buffer = make([]byte, 0)
	return state.wire.Close()
}

func (state KiSS_State) Read(p []byte) (int, error) {
	// Return anything in buffer first!
	if len(*state.buffer) > 0 {
		n := copy(p, *state.buffer)
		*state.buffer = (*state.buffer)[n:]
		fmt.Println("Have to make do mit buffer!")
		return n, nil
	}

	defer func() { state.read_packets++ }()
	nonce := make([]byte, 8)
	binary.BigEndian.PutUint64(nonce, state.read_packets)
	segment, err := KiSS_read_segment(state.wire)
	//LOG(LOG_DEBUG, "Obtained segment: %s|%X", segment.StringRep(), segment.Bytes())
	if err != nil {
		return 0, err
	}

	// now must of decode da segment
	if segment.segment_type == K_APP_DATA {
		// Application data
		rawdat := segment.raw_payload
		toret, err := state.read_ciph.Open(rawdat)
		check_fatal(err)
		FASSERT(len(toret) <= len(p))
		// Now we must buffer.
		if len(toret) > len(p) {
			*state.buffer = append(*state.buffer, toret[len(p):]...)
			copy(p, toret)
		} else {
			copy(p, toret)
		}
		return len(p), err
	} else {
		SPANIC("Alerts not implemented yet!")
	}
	SPANIC("WTF?!?!?!?!?")
	return 0, io.EOF
}
