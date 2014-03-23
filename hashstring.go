// hashstring.go
package main

import (
	"encoding/base32"
	"kirisurf/ll/kicrypt"
	"strings"
)

func hash_base32(data []byte) string {
	return strings.ToLower(base32.StdEncoding.EncodeToString(
		kicrypt.InvariantHash(data)[:20]))
}
