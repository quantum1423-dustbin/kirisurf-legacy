package kiss

// #include "cfuncs.h"
import "C"
import "unsafe"

type tea_state struct {
	key []byte
}

func (thing *tea_state) Decrypt(dst, src []byte) {
	panic("Wtf....")
}

func (thing *tea_state) BlockSize() int {
	return 8
}

func (thing *tea_state) Encrypt(dst, src []byte) {
	copy(dst, src)
	vptr := (*C.uint32_t)(unsafe.Pointer(&dst[0]))
	kptr := (*C.uint32_t)(unsafe.Pointer(&thing.key[0]))
	C.tea_encrypt(vptr, kptr)
}

func make_tea(key []byte) *tea_state {
	xaxa := new(tea_state)
	xaxa.key = key
	return xaxa
}
