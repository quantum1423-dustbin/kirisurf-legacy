package kiss

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"testing"
	"unsafe"
	"code.google.com/p/go.crypto/ripemd160"

	"github.com/codahale/chacha20"
	"github.com/dchest/blake256"
)

func BenchmarkChacha20(b *testing.B) {
	val := make([]byte, 1024)
	key := make([]byte, 32)
	xaxa, _ := chacha20.NewCipher(key, make([]byte, 8))
	for i := 0; i < b.N; i++ {
		xaxa.XORKeyStream(val, val)
	}
}

func BenchmarkChugger(b *testing.B) {
	val := make([]byte, 1024)
	var key = make([]byte, 56)
	gaga := make_chugger(key)
	for i := 0; i < b.N; i++ {
		gaga.Seal(val)
	}
}

func TestChugger(t *testing.T) {
	var key = make([]byte, 32)
	gaga := make_chugger(key)
	dada := make_chugger(key)
	pt := []byte("Hello world!")
	for i := 0; i < 100; i++ {
		ct := gaga.Seal(pt)
		pt2, err := dada.Open(ct)
		if err != nil {
			panic(err)
		}
		if string(pt) != string(pt2) {
			t.Fail()
		}
	}
}

func TestUnsafe(t *testing.T) {
	xaxa := uint64(0x0102030405060708)
	xaxaptr := unsafe.Pointer(&xaxa)
	xaxaarr := (*[8]byte)(xaxaptr)
	fmt.Println(xaxaarr)
}

func BenchmarkBLAKE256(b *testing.B) {
	val := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		xaxa := blake256.New()
		xaxa.Write(val)
		xaxa.Sum(nil)
	}
}

func BenchmarkSHA1HMAC(b *testing.B) {
	val := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		yuyu := hmac.New(sha1.New, make([]byte, 16))
		yuyu.Write(val)
		yuyu.Sum(nil)
	}
}

func BenchmarkRIPEMD160HMAC(b *testing.B) {
	val := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		yuyu := hmac.New(ripemd160.New, make([]byte, 16))
		yuyu.Write(val)
		yuyu.Sum(nil)
	}
}
