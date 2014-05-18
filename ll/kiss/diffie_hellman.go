package kiss

import (
	"crypto/rand"
	"math/big"
)

type dh_public_key []byte
type dh_private_key []byte

type DHKeys struct {
	Public  dh_public_key
	Private dh_private_key
}

func dh_gen_secret(our_private dh_private_key, their_public dh_public_key) []byte {
	bitlen := len(our_private) * 8
	// checks
	if bitlen != 1536 && bitlen != 2048 {
		panic("Why are you trying to generate DH key with wrong bitlen?")
	}
	var group *big.Int
	if bitlen == 1536 {
		group = dh_group_5
	} else {
		group = dh_group_14
	}
	return big.NewInt(0).Exp(big.NewInt(0).SetBytes(their_public),
		big.NewInt(0).SetBytes(our_private), group).Bytes()
}

func GenerateDHKeys() DHKeys {
	return dh_gen_key(2048)
}

func dh_gen_key(bitlen int) DHKeys {
	// checks
	if bitlen != 1536 && bitlen != 2048 {
		panic("Why are you trying to generate DH key with wrong bitlen?")
	}
	var group *big.Int
	if bitlen == 1536 {
		group = dh_group_5
	} else {
		group = dh_group_14
	}

	// randomly generate even private key
	pub := dh_public_key(make([]byte, bitlen/8))
	priv := dh_private_key(make([]byte, bitlen/8))
	rand.Read(priv)
	priv[bitlen/8-1] /= 2
	priv[bitlen/8-1] *= 2
	privBnum := big.NewInt(0).SetBytes(priv)

retry:
	// generate public key
	pubBnum := big.NewInt(0).Exp(big.NewInt(2), privBnum, group)
	ggg := make([]byte, 1)
	rand.Read(ggg)
	if ggg[0]%2 == 0 {
		pubBnum = big.NewInt(0).Sub(group, pubBnum)
	}

	// Obtain pubkey
	candid := pubBnum.Bytes()
	if len(candid) != len(pub) {
		goto retry
	}
	copy(pub, candid)

	return DHKeys{pub, priv}
}
