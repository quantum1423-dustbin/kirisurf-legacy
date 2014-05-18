package kiss

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"fmt"

	"code.google.com/p/go.crypto/salsa20"
)

// This file implements the Grand Central Chugger, which handles stream authentication
// Salsa20+HMAC-SHA1 is used.

type chugger struct {
	key *[32]byte
}

func (ctx *chugger) Seal(pt []byte) []byte {
	toret := make([]byte, 24+20+len(pt))
	rand.Read(toret[:24])
	nonce := toret[:24]

	xaxa := hmac.New(sha1.New, ctx.key[:])
	xaxa.Write(pt)
	tag := xaxa.Sum(nil)
	pt = append(tag, pt...)

	salsa20.XORKeyStream(toret[24:], pt, nonce, ctx.key)
	fmt.Printf("sealing %x to %x\n", pt, toret)
	return toret
}

func (ctx *chugger) Open(ct []byte) ([]byte, error) {
	if len(ct) < 24+20 {
		return nil, ErrPacketTooShort
	}
	nonce := ct[:24]
	oct = ct
	ct = ct[24:]
	pt := make([]byte, len(ct))
	salsa20.XORKeyStream(pt, ct, nonce, ctx.key)
	xaxa := hmac.New(sha1.New, ctx.key[:])
	xaxa.Write(pt[20:])
	actual_sum := xaxa.Sum(nil)
	hypo_sum := pt[:20]

	if subtle.ConstantTimeCompare(actual_sum, hypo_sum) == 1 {
		fmt.Printf("opening %x to %x\n", oct, pt)
		return pt[20:], nil
	}
	return nil, ErrMacNoMatch
}
