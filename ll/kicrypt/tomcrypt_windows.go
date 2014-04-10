package kicrypt

import (
	"crypto/cipher"
	"crypto/subtle"
	"errors"
	"fmt"
	"hash"
	"kirisurf/ll/buf"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"
)

// #cgo LDFLAGS: libtomcrypt.win32.a
// #cgo CFLAGS: -I ./tomcrypt_headers
// #include "./tomcrypt_headers/tomcrypt.h"
import "C"

func FASSERT(cond bool) {
	_, file, line, _ := runtime.Caller(1)
	if !cond {
		fmt.Fprintf(os.Stderr, "**** KURWA!!! ****\n")
		fmt.Fprintf(os.Stderr, "Error encountered was of ASSERTION. Kirisurf cannot into continue. Gib bugfix fast pl0x.\n")
		fmt.Fprintf(os.Stderr, "Bad FASSERT called in %s:%d\n", filepath.Base(file), line)
		panic("bad assert")
	}
}

type fastAES struct {
	schedule []byte
}

type fastSHA512State struct {
	scratch []byte
}

func (bloo fastSHA512State) Write(inp []byte) (n int, e error) {
	C.sha512_process((*C.hash_state)(unsafe.Pointer(&bloo.scratch[0])),
		unsafe_bytes(inp), C.ulong(len(inp)))
	return len(inp), nil
}

func (bloo fastSHA512State) Reset() {
	abla := make([]byte, 512/8)
	C.sha512_done((*C.hash_state)(unsafe.Pointer(&bloo.scratch[0])),
		unsafe_bytes(abla))
}

func (bloo fastSHA512State) Size() int {
	return 512 / 8
}

func (bloo fastSHA512State) Sum(goo []byte) []byte {
	abla := make([]byte, 512/8)
	C.sha512_done((*C.hash_state)(unsafe.Pointer(&bloo.scratch[0])),
		unsafe_bytes(abla))
	buf.Free(bloo.scratch)
	return append(goo, abla...)
}

func (bloo fastSHA512State) BlockSize() int {
	return 128
}

func fastSHA512() hash.Hash {
	scratch := buf.Alloc()
	C.sha512_init((*C.hash_state)(unsafe.Pointer(&scratch[0])))
	goo := fastSHA512State{scratch}
	return hash.Hash(goo)
}

func (sch fastAES) Encrypt(dst, src []byte) {
	C.aes_ecb_encrypt(unsafe_bytes(src), unsafe_bytes(dst),
		(*C.symmetric_key)(unsafe.Pointer(&sch.schedule[0])))
}

func (sch fastAES) Decrypt(dst, src []byte) {
	C.aes_ecb_decrypt(unsafe_bytes(src), unsafe_bytes(dst),
		(*C.symmetric_key)(unsafe.Pointer(&sch.schedule[0])))
}

func (sch fastAES) BlockSize() int {
	return 16
}

func fastAES_initialize(key []byte) cipher.Block {
	if !(len(key) == 16 || len(key) == 32) {
		panic("AES must use 128 or 256 bits in a key")
	}
	aes_schedule := make([]byte, 4096)
	C.aes_setup((*C.uchar)(unsafe.Pointer(&(key[0]))), C.int(len(key)), 0,
		(*C.symmetric_key)(unsafe.Pointer(&aes_schedule[0])))
	toret := cipher.Block(fastAES{aes_schedule})
	return toret
}

func unsafe_bytes(bts []byte) *C.uchar {
	return (*C.uchar)(unsafe.Pointer(&bts[0]))
}

type fastGCMState []byte

func (state fastGCMState) NonceSize() int {
	return 12
}

func (state fastGCMState) Overhead() int {
	panic("WTF!!!?!?!?!???!????!?!?")
	return -1
}

func (state fastGCMState) Seal(dst, nonce, plaintext, data []byte) []byte {
	rawenc := make([]byte, len(plaintext))
	sched := (*_Ctype_gcm_state)(unsafe.Pointer(&state[0]))
	FASSERT(C.gcm_reset(sched) == C.CRYPT_OK)
	FASSERT(C.gcm_add_iv(sched, unsafe_bytes(nonce), 12) == C.CRYPT_OK)
	C.gcm_add_aad(sched, nil, 0)
	FASSERT(C.gcm_process(sched, unsafe_bytes(plaintext), C.ulong(len(plaintext)),
		unsafe_bytes(rawenc), C.GCM_ENCRYPT) == C.CRYPT_OK)
	tag := buf.Alloc()
	thing := C.ulong(16)
	FASSERT(C.gcm_done(sched, unsafe_bytes(tag), &thing) == C.CRYPT_OK)
	rawenc = append(rawenc, tag[:int(thing)]...)
	buf.Free(tag)
	return append(dst, rawenc...)
}

func (state fastGCMState) Open(dst, nonce, ciphertext, data []byte) ([]byte, error) {
	rawpt := make([]byte, len(ciphertext)-16)
	sched := (*_Ctype_gcm_state)(unsafe.Pointer(&state[0]))
	FASSERT(C.gcm_reset(sched) == C.CRYPT_OK)
	FASSERT(C.gcm_add_iv(sched, unsafe_bytes(nonce), 12) == C.CRYPT_OK)
	FASSERT(C.gcm_add_aad(sched, nil, 0) == C.CRYPT_OK)
	FASSERT(C.gcm_process(sched, unsafe_bytes(rawpt), C.ulong(len(rawpt)),
		unsafe_bytes(ciphertext), C.GCM_DECRYPT) == C.CRYPT_OK)
	tag := buf.Alloc()
	thing := C.ulong(16)
	C.gcm_done(sched, unsafe_bytes(tag), &thing)

	if subtle.ConstantTimeCompare(tag[:int(thing)], ciphertext[len(rawpt):]) != 1 {
		return nil, errors.New("WTF! HASH OF MISMATCHINGS!")
	}
	buf.Free(tag)
	return append(dst, rawpt...), nil
}

func fastAES_GCM(rwkey []byte) fastGCMState {
	key := hash_invar(rwkey)[:16]
	state := make([]byte, 65536*20)
	sched := (*_Ctype_gcm_state)(unsafe.Pointer(&state[0]))
	idx := C.find_cipher(C.CString("aes"))
	FASSERT(C.gcm_init(sched, idx, unsafe_bytes(key), C.int(len(key))) == C.CRYPT_OK)
	//LOG(LOG_DEBUG, "%X", state)
	return fastGCMState(state)
}

type fastBF_State []byte

func fastBF_NewOFB(key []byte) fastBF_State {
	idx := C.find_cipher(C.CString("blowfish"))
	iv := make([]byte, 8)
	state := make([]byte, 65536)
	FASSERT(C.ofb_start(idx,
		unsafe_bytes(iv),
		unsafe_bytes(key),
		C.int(len(key)),
		C.int(0), (*_Ctype_symmetric_OFB)((unsafe.Pointer)(&state[0]))) == C.CRYPT_OK)
	return state
}

func (state fastBF_State) XORKeyStream(dst, src []byte) {
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
}
