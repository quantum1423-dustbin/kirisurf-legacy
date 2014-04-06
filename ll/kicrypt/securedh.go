// securedh.go
package kicrypt

import (
	"crypto/rand"
	"encoding/pem"
	big "github.com/ncw/gmp"
	"io/ioutil"
)

var GROUP14 = func() *big.Int {
	toret := big.NewInt(0)
	toret.SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D2261898FA051015728E5A8AACAA68FFFFFFFFFFFFFFFF", 16)
	return toret
}()

type SecureDH_keypair struct {
	Private *big.Int
	Public  *big.Int
}

func SecureDH_genpair() SecureDH_keypair {
	//generate 2048-bit (256-byte) random private key
	private := make([]byte, 256)
	rand.Read(private)
	privateInt := big.NewInt(0).SetBytes(private)
	publicInt := big.NewInt(0).Exp(big.NewInt(2), privateInt, GROUP14)
	if len(privateInt.Bytes()) == 256 && len(privateInt.Bytes()) == 256 {
		return SecureDH_keypair{privateInt, publicInt}
	} else {
		return SecureDH_genpair()
	}
}

func SecureDH_gensecret(our_private, their_public *big.Int) []byte {
	return big.NewInt(0).Exp(their_public, our_private, GROUP14).Bytes()
}

func (pr SecureDH_keypair) SavePrivate(fname string) {
	blk := pem.Block{"SecureDH PRIVATE KEY", nil, pr.Private.Bytes()}
	err := ioutil.WriteFile(fname, pem.EncodeToMemory(&blk), 600)
	if err != nil {
		panic("Problem saving private key")
	}
}

func SecureDH_loadkey(fname string) SecureDH_keypair {
	bts, err := ioutil.ReadFile(fname)
	if err != nil {
		panic("Problem reading private key")
	}
	p, _ := pem.Decode(bts)
	privkey := big.NewInt(0).SetBytes(p.Bytes)
	pubkey := big.NewInt(0).Exp(big.NewInt(2), privkey, GROUP14)
	return SecureDH_keypair{privkey, pubkey}
}
