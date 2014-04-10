package main

import (
	"crypto/subtle"
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
	"math/big"
	"net"
)

type Subcircuit struct {
	cpath []dirclient.KNode
	wire  io.ReadWriteCloser
}

func build_subcircuit() (*Subcircuit, error) {
	slc := dirclient.FindPath(MasterConfig.Network.MinCircuitLen)
	DEBUG("Building a subcicruit with minlen %d...", len(slc))
	// this returns a checker whether a public key is valid
	pubkey_checker := func(hsh string) kiss.Verifier {
		return func(k *big.Int) bool {
			hashed := hash_base32(k.Bytes())
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
		DEBUG("Extending circuit to length %d...", idx+2)
		// extend wire
		err = write_sc_message(sc_message{SC_EXTEND, ele.PublicKey}, wire)
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
		DEBUG("Extended circuit to length %d", idx+2)
	}
	err = write_sc_message(sc_message{SC_TERMINATE, ""}, wire)
	if err != nil {
		return nil, err
	}
	DEBUG("Subcircuit building complete with length %d", len(slc))
	toret := Subcircuit{slc, wire}
	return &toret, nil
}
