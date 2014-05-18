package kiss

import "encoding/binary"

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
	_v := make([]byte, 8)
	copy(_v, src)
	_k := thing.key
	var v [2]uint32
	var k [4]uint32
	v[0] = binary.BigEndian.Uint32(_v[0:4])
	v[1] = binary.BigEndian.Uint32(_v[4:8])
	k[0] = binary.BigEndian.Uint32(_k[0:4])
	k[1] = binary.BigEndian.Uint32(_k[4:8])
	k[2] = binary.BigEndian.Uint32(_k[8:12])
	k[3] = binary.BigEndian.Uint32(_k[12:16])

	v0 := v[0]
	v1 := v[1]
	delta := uint32(0x9e3779b9)
	k0 := k[0]
	k1 := k[1]
	k2 := k[2]
	k3 := k[3]
	sum := uint32(0)

	for i := 0; i < 12; i++ {
		sum += delta
		v0 += ((v1 << 4) + k0) ^ (v1 + sum) ^ ((v1 >> 5) + k1)
		v1 += ((v0 << 4) + k2) ^ (v0 + sum) ^ ((v0 >> 5) + k3)
	}

	binary.BigEndian.PutUint32(_v[0:4], v0)
	binary.BigEndian.PutUint32(_v[4:8], v1)

	copy(dst, _v)
}

func make_tea(key []byte) *tea_state {
	xaxa := new(tea_state)
	xaxa.key = key
	return xaxa
}
