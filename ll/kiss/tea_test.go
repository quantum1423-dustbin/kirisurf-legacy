package kiss

import (
	"crypto/cipher"
	"testing"
)

/*
func TestTEA(t *testing.T) {
	val := make([]byte, 8)
	key := make([]byte, 16)
	tea_encrypt(val, key)
	fmt.Printf("%x\n", val)
	val = make([]byte, 8)
	key = make([]byte, 16)
	tea_encrypt(val, key)
	fmt.Printf("%x\n", val)
}*/

func BenchmarkTEA(b *testing.B) {
	val := make([]byte, 8)
	key := make([]byte, 16)
	xaxa := make_tea(key)
	cc := cipher.NewCTR(xaxa, val)
	ddd := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		cc.XORKeyStream(ddd, ddd)
	}
}
