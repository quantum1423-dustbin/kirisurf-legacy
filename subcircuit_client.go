package main

import (
	"crypto/subtle"
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
	"math/big"
	"net"

	"github.com/coreos/go-log/log"
)

type Subcircuit struct {
	cpath []dirclient.KNode
	wire  io.ReadWriteCloser
}

func build_subcircuit() (*Subcircuit, error) {
	log.Debug("Into buildings sc")
	slc := dirclient.FindPath(MasterConfig.Network.MinCircuitLen)
	log.Debug(slc)
	// this returns a checker whether a public key is valid
	pubkey_checker := func(hsh string) kiss.Verifier {
		return func(k *big.Int) bool {
			hashed := hash_base32(k.Bytes())
			log.Debugf("Comparing %s with %s", hsh, hashed)
			return subtle.ConstantTimeCompare([]byte(hashed), []byte(hsh)) == 1
		}
	}
	// circuit-building loop
	iwire, err := net.Dial("tcp", slc[0].Address)
	if err != nil {
		return nil, err
	}
	iwire, err = kiss.Kiriobfs_handshake_client(iwire)
	if err != nil {
		//iwire.Close()
		return nil, err
	}
	wire, err := kiss.KiSS_handshake_client(
		iwire, pubkey_checker(slc[0].PublicKey))
	if err != nil {
		//wire.Close()
		return nil, err
	}
	for idx, ele := range slc[1:] {
		log.Debug(idx)
		// extend wire
		err = write_sc_message(sc_message{SC_EXTEND, ele.PublicKey}, wire)
		log.Debug(ele.PublicKey)
		if err != nil {
			wire.Close()
			return nil, err
		}

		verifier := pubkey_checker(ele.PublicKey)
		// at this point wire is raw (well unobfs) connection to next
		wire, err = kiss.KiSS_handshake_client(wire, verifier)
		if err != nil {
			//wire.Close()
			return nil, err
		}
		log.Debug("Of connected into sc %d", idx)
	}
	err = write_sc_message(sc_message{SC_TERMINATE, ""}, wire)
	if err != nil {
		return nil, err
	}
	log.Debug("Yay subcircuit connectings of dones.")
	toret := Subcircuit{slc, wire}
	return &toret, nil
}
