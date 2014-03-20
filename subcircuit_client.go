package main

import (
	"crypto/subtle"
	"fmt"
	"io"
	"libkiridir"
	"libkiss"
	"math/big"
	"net"

	"github.com/coreos/go-log/log"
)

type Subcircuit struct {
	cpath []libkiridir.KNode
	wire  io.ReadWriteCloser
}

func build_subcircuit() (*Subcircuit, error) {
	lst := libkiridir.FindPath(MasterConfig.Network.MinCircuitLen)
	slc := lst.ToSlice()
	log.Debug(slc)
	// this returns a checker whether a public key is valid
	pubkey_checker := func(hsh string) libkiss.Verifier {
		return func(k *big.Int) bool {
			hashed := hash_base32(k.Bytes())
			return subtle.ConstantTimeCompare([]byte(hashed), []byte(hsh)) == 1
		}
	}
	// circuit-building loop
	iwire, err := net.Dial("tcp", slc[0].Address)
	if err != nil {
		iwire.Close()
		return nil, err
	}
	iwire, err = libkiss.Kiriobfs_handshake_client(iwire)
	if err != nil {
		iwire.Close()
		return nil, err
	}
	wire, err := libkiss.KiSS_handshake_client(
		iwire, pubkey_checker(slc[0].PublicKey))
	if err != nil {
		wire.Close()
		return nil, err
	}
	for idx, ele := range slc[1:] {
		// extend wire
		_, err = fmt.Fprintf(wire, "CONN %s\n", ele.PublicKey)
		if err != nil {
			wire.Close()
			return nil, err
		}
		verifier := pubkey_checker(ele.PublicKey)
		// at this point wire is raw (well unobfs) connection to next
		wire, err = libkiss.KiSS_handshake_client(wire, verifier)
		if err != nil {
			//wire.Close()
			return nil, err
		}
		log.Debug("Of connected into sc %d", idx)
	}
	log.Debug("Yay subcircuit connectings of dones.")
	panic("Cannot into lives, can only into dyings.")
}
