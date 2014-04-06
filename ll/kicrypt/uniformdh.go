// uniformdh.go
package kicrypt

import (
	"crypto/rand"
	big "github.com/ncw/gmp"
)

var GROUP5 = func() *big.Int {
	toret := big.NewInt(0)
	toret.SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA237327FFFFFFFFFFFFFFFF", 16)
	return toret
}()

type UniformDH_keypair struct {
	Private *big.Int
	Public  *big.Int
}

func (kp UniformDH_keypair) PrivateBytes() []byte {
	return kp.Private.Bytes()
}

func (kp UniformDH_keypair) PublicBytes() []byte {
	return kp.Public.Bytes()
}

func UniformDH_genpair() UniformDH_keypair {
	//generate 1536-bit (192-byte) *even* random private key
	private := make([]byte, 192)
	rand.Read(private)
	privateInt := big.NewInt(0).SetBytes(private)
	privateInt.Div(privateInt, big.NewInt(2))
	privateInt.Mul(privateInt, big.NewInt(2))

	publicInt := big.NewInt(0).Exp(big.NewInt(2), privateInt, GROUP5)
	ggg := make([]byte, 1)
	_, err := rand.Read(ggg)
	ggg[0] = ggg[0] % 2
	if err != nil {
		panic(err.Error())
	}
	if ggg[0] == 0 {
		publicInt = big.NewInt(0).Sub(GROUP5, publicInt)
	}
	if len(privateInt.Bytes()) == 192 && len(publicInt.Bytes()) == 192 {
		return UniformDH_keypair{privateInt, publicInt}
	} else {
		return UniformDH_genpair()
	}
}

func UniformDH_gensecret(our_private, their_public *big.Int) []byte {
	return big.NewInt(0).Exp(their_public, our_private, GROUP5).Bytes()
}
