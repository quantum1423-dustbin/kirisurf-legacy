package kiss

import "encoding/binary"

type tea_state struct {
	k0 uint32
	k1 uint32
	k2 uint32
	k3 uint32
}

func (thing *tea_state) Decrypt(dst, src []byte) {
	panic("Wtf....")
}

func (thing *tea_state) BlockSize() int {
	return 8
}

func (thing *tea_state) Encrypt(dst, src []byte) {
	v0 := binary.LittleEndian.Uint32(src[0:4])
	v1 := binary.LittleEndian.Uint32(src[4:8])
	delta := uint32(0x9e3779b9)
	k0 := thing.k0
	k1 := thing.k1
	k2 := thing.k2
	k3 := thing.k3
	sum := uint32(0)

	for i := 0; i < 16; i++ {
		sum += delta
		v0 += ((v1 << 4) + k0) ^ (v1 + sum) ^ ((v1 >> 5) + k1)
		v1 += ((v0 << 4) + k2) ^ (v0 + sum) ^ ((v0 >> 5) + k3)
	}

	binary.LittleEndian.PutUint32(dst[0:4], v0)
	binary.LittleEndian.PutUint32(dst[4:8], v1)
}

func make_tea(key []byte) *tea_state {
	xaxa := new(tea_state)
	xaxa.k0 = binary.LittleEndian.Uint32(key[0:4])
	xaxa.k1 = binary.LittleEndian.Uint32(key[4:8])
	xaxa.k2 = binary.LittleEndian.Uint32(key[8:12])
	xaxa.k3 = binary.LittleEndian.Uint32(key[12:16])
	return xaxa
}
