package main

import (
	"crypto/subtle"
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
	"net"

	"github.com/KirisurfProject/kilog"
)

func build_subcircuit(slc []dirclient.KNode) (io.ReadWriteCloser, error) {
	// this returns a checker whether a public key is valid
	pubkey_checker := func(hsh string) func([]byte) bool {
		return func([]byte) bool { return true }

		return func(xaxa []byte) bool {
			hashed := hash_base32(xaxa)
			return subtle.ConstantTimeCompare([]byte(hashed), []byte(hsh)) == 1
		}
	}
	// circuit-building loop
	iwire, err := net.Dial("tcp", slc[0].Address)
	if err != nil {
		return nil, err
	}
	gwire, err := kiss.Obfs3fHandshake(iwire, false)
	if err != nil {
		iwire.Close()
		return nil, err
	}
	wire, err := kiss.TransportHandshake(kiss.GenerateDHKeys(),
		gwire, pubkey_checker(slc[0].PublicKey))
	if err != nil {
		iwire.Close()
		return nil, err
	}
	for _, ele := range slc[1:] {
		// extend wire
		_, err = wire.Write(append([]byte{byte(len(ele.PublicKey))}, ele.PublicKey...))
		if err != nil {
			iwire.Close()
			return nil, err
		}

		verifier := pubkey_checker(ele.PublicKey)
		// at this point wire is raw (well unobfs) connection to next
		wire, err := kiss.Obfs3fHandshake(wire, false)
		if err != nil {
			kilog.Debug("Died when obfs3 at %s", ele.PublicKey)
			iwire.Close()
			return nil, err
		}
		wire, err = kiss.TransportHandshake(kiss.GenerateDHKeys(), wire, verifier)
		if err != nil {
			kilog.Debug("Died when transport at %s", ele.PublicKey)
			iwire.Close()
			return nil, err
		}
		kilog.Debug("Connected to %v", ele)
	}
	_, err = wire.Write([]byte("\000"))
	if err != nil {
		iwire.Close()
		return nil, err
	}
	kilog.Debug("Sent SC_TERMINATE")
	return wire, nil
}
