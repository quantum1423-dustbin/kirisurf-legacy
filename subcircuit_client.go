package main

import (
	"crypto/subtle"
	"fmt"
	"io"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/intercom"
	"kirisurf/ll/kiss"
	"strconv"
	"strings"

	"github.com/KirisurfProject/kilog"
)

func old2new(addr string) string {
	port, _ := strconv.Atoi(strings.Split(addr, ":")[1])
	naddr := fmt.Sprintf("kirisurf@%s:%d", strings.Split(addr, ":")[0], port+1)
	return naddr
}

var dialer *intercom.IntercomDialer

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
	gwire, err := dialer.Dial(old2new(slc[0].Address))
	if err != nil {
		gwire.Close()
		return nil, err
	}
	wire, err := kiss.TransportHandshake(kiss.GenerateDHKeys(),
		gwire, pubkey_checker(slc[0].PublicKey))
	if err != nil {
		gwire.Close()
		return nil, err
	}
	for _, ele := range slc[1:] {
		// extend wire
		_, err = wire.Write(append([]byte{byte(len(ele.PublicKey))}, ele.PublicKey...))
		if err != nil {
			gwire.Close()
			return nil, err
		}

		verifier := pubkey_checker(ele.PublicKey)
		// at this point wire is raw (well unobfs) connection to next
		wire, err = kiss.TransportHandshake(kiss.GenerateDHKeys(), wire, verifier)
		if err != nil {
			kilog.Debug("Died when transport at %s", ele.PublicKey)
			gwire.Close()
			return nil, err
		}
		kilog.Debug("Connected to %v", ele)
	}
	_, err = wire.Write([]byte("\000"))
	if err != nil {
		gwire.Close()
		return nil, err
	}
	kilog.Debug("Sent SC_TERMINATE")
	return wire, nil
}

func init() {
	dialer = intercom.MakeIntercomDialer()
}
