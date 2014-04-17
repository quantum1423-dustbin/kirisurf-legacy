// +build cgo

package kicrypt

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
)

// #cgo LDFLAGS: libtomcrypt.amd64.a
// #cgo CFLAGS: -I ./tomcrypt_headers
// #include "./tomcrypt_headers/tomcrypt.h"
import "C"

func unsafe_bytes(bts []byte) *C.uchar {
	return (*C.uchar)(unsafe.Pointer(&bts[0]))
}

type fastTF_State []byte

func FASSERT(cond bool) {
	_, file, line, _ := runtime.Caller(1)
	if !cond {
		fmt.Fprintf(os.Stderr, "**** KURWA!!! ****\n")
		fmt.Fprintf(os.Stderr, "Error encountered was of ASSERTION. Kirisurf cannot into continue. Gib bugfix fast pl0x.\n")
		fmt.Fprintf(os.Stderr, "Bad FASSERT called in %s:%d\n", filepath.Base(file), line)
		panic("bad assert")
	}
}

func fastTF_NewOFB(key, iv []byte) fastTF_State {
	idx := C.find_cipher(C.CString("aes"))
	state := make([]byte, 65536)
	FASSERT(C.ofb_start(idx,
		unsafe_bytes(iv),
		unsafe_bytes(key),
		C.int(len(key)),
		C.int(0), (*_Ctype_symmetric_OFB)((unsafe.Pointer)(&state[0]))) == C.CRYPT_OK)
	return state
}

func (state fastTF_State) XORKeyStream(dst, src []byte) {
	C.ofb_encrypt(unsafe_bytes(src), unsafe_bytes(dst), C.ulong(len(src)),
		(*_Ctype_symmetric_OFB)((unsafe.Pointer)(&state[0])))
}

var sha512idx C.int

func fastHMAC(msg, key []byte) []byte {
	toret := make([]byte, 512/8)
	thing := C.ulong(512 / 8)
	C.hmac_memory(sha512idx,
		unsafe_bytes(key), C.ulong(len(key)),
		unsafe_bytes(msg), C.ulong(len(msg)),
		unsafe_bytes(toret), (&thing))
	return toret
}

func init() {
	idx := C.register_cipher(&C.aes_desc)
	FASSERT(idx != -1)
	sha512idx = C.register_hash(&C.sha512_desc)
	FASSERT(sha512idx != -1)

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
