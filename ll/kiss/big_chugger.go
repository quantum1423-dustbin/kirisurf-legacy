package kiss

import (
	"crypto/cipher"
	"crypto/subtle"
	"encoding/binary"

	"github.com/codahale/chacha20"
	"github.com/dchest/blake2s"
)

// This file implements the Grand Central Chugger, which handles stream authentication
// Stream cipher ChaCha20 is used.

type chugger struct {
	streamer  cipher.Stream
	key       []byte
	read_num  uint64
	write_num uint64
}

func (ctx *chugger) Seal(pt []byte) []byte {
	seq := make([]byte, 8)
	binary.LittleEndian.PutUint64(seq, ctx.write_num)
	ctx.write_num++

	toret := make([]byte, 20+len(pt))

	xaxa := blake2s.NewMAC(20, ctx.key)
	xaxa.Write(pt)
	xaxa.Write(seq)
	tag := xaxa.Sum(nil)
	pt = append(tag, pt...)

	ctx.streamer.XORKeyStream(toret, pt)
	return toret
}

func (ctx *chugger) Open(ct []byte) ([]byte, error) {
	if len(ct) < 20 {
		return nil, ErrPacketTooShort
	}
	seq := make([]byte, 8)
	binary.LittleEndian.PutUint64(seq, ctx.read_num)
	ctx.read_num++

	pt := make([]byte, len(ct))
	ctx.streamer.XORKeyStream(pt, ct)
	xaxa := blake2s.NewMAC(20, ctx.key)
	xaxa.Write(pt[20:])
	xaxa.Write(seq)
	actual_sum := xaxa.Sum(nil)
	hypo_sum := pt[:20]

	if subtle.ConstantTimeCompare(actual_sum, hypo_sum) == 1 {
		return pt[20:], nil
	}
	return nil, ErrMacNoMatch
}

func make_chugger(key []byte) *chugger {
	state, err := chacha20.NewCipher(key, make([]byte, 8))
	if err != nil {
		panic(err.Error())
	}
	return &chugger{state, key, 0, 0}
}
