// hashstring.go
package main

import (
	"encoding/base32"
	"libkiricrypt"
	"strings"
)

func hash_base32(data []byte) string {
	return strings.ToLower(base32.StdEncoding.EncodeToString(
		libkiricrypt.InvariantHash(data)[:20]))
}
