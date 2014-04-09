// +build !cgo

package kicrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha512"
	"fmt"
	"os"
)

func FASSERT(cond bool) {
	if !cond {
		fmt.Fprintf(os.Stderr, "**** KURWA!!! ****\n")
		fmt.Fprintf(os.Stderr, "Error encountered was of ASSERTION. Kirisurf cannot into continue. Gib bugfix fast pl0x.\n")
		fmt.Fprintf(os.Stderr, "Bad FASSERT called in")
		panic("bad assert")
	}
}

func fastAES_GCM(rwkey []byte) cipher.AEAD {
	key := hash_invar(rwkey)[:16]
	ciph, _ := aes.NewCipher(key)
	s, _ := cipher.NewGCM(ciph)
	return s
}

func fastHMAC(msg, key []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(msg)
	return h.Sum(nil)
}

func fastAES_initialize(aaa ...interface{}) cipher.Block {
	panic("Not supposed to be called -_-")
}

func fastBF_NewOFB(aaa ...interface{}) cipher.Stream {
	panic("Not supposed to be called -_-")
}

func init() {
	panic("NoC implementation NOT DONE YET!")
}
