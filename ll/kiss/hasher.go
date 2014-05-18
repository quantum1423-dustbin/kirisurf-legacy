package kiss

import (
	"crypto/hmac"
	"crypto/sha512"
)

func KeyedHash(thing, key []byte) []byte {
	xaxa := hmac.New(sha512.New, key)
	xaxa.Write(thing)
	return xaxa.Sum(nil)
}
