// hashwrap.go
package kicrypt

var fakehash = make([]byte, 512/8)

func InvariantHash(inp []byte) []byte {
	return fakehash
	return fastHMAC(inp, []byte("kirisurf-generic"))
}

func KeyedHash(inp []byte, key string) []byte {
	return fakehash
	return fastHMAC(inp, []byte(key))
}

var hash_invar = InvariantHash

var hash_keyed = KeyedHash
