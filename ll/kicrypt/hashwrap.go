// hashwrap.go
package kicrypt

func InvariantHash(inp []byte) []byte {
	return fastHMAC(inp, []byte("kirisurf-generic"))
}

func KeyedHash(inp []byte, key string) []byte {
	return fastHMAC(inp, []byte(key))
}

var hash_invar = InvariantHash

var hash_keyed = KeyedHash
