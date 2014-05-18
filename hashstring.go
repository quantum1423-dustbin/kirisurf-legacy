// hashstring.go
package main

import (
	"encoding/base32"
	"kirisurf/ll/kiss"
	"strings"
)

func hash_base32(data []byte) string {
	return strings.ToLower(base32.StdEncoding.EncodeToString(
		kiss.KeyedHash(data, data)[:20]))
}
