package kicrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func FASSERT(cond bool) {
	_, file, line, _ := runtime.Caller(1)
	if !cond {
		fmt.Fprintf(os.Stderr, "**** KURWA!!! ****\n")
		fmt.Fprintf(os.Stderr, "Error encountered was of ASSERTION. Kirisurf cannot into continue. Gib bugfix fast pl0x.\n")
		fmt.Fprintf(os.Stderr, "Bad FASSERT called in %s:%d\n", filepath.Base(file), line)
		panic("bad assert")
	}
}

func fastTF_NewOFB(key, iv []byte) cipher.Stream {
	thing, _ := aes.NewCipher(key)
	ofb := cipher.NewOFB(thing, iv)
	return ofb
}

func fastHMAC(msg, key []byte) []byte {
	h := hmac.New(sha512.New, key)
	h.Write(msg)
	return h.Sum(nil)
}

func init() {
	// Test vectors
	key, _ := hex.DecodeString("603deb1015ca71be2b73aef0857d77811f352c073b6108d72d9810a30914dff4")
	iv, _ := hex.DecodeString("B7BF3A5DF43989DD97F0FA97EBCE2F4A")
	pt, _ := hex.DecodeString("ae2d8a571e03ac9c9eb76fac45af8e51")
	ct := make([]byte, len(pt))
	state := fastTF_NewOFB(key, iv)
	state.XORKeyStream(ct, pt)
	if hex.EncodeToString(ct) != "4febdc6740d20b3ac88f6ad82a4fb08d" {
		panic(fmt.Sprintf("AES test returned %s, should be %s", hex.EncodeToString(ct),
			"4febdc6740d20b3ac88f6ad82a4fb08d"))
	}
}
