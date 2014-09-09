package kiridht

import "encoding/base32"

type dhtkey []byte

func (lel dhtkey) tostring() string {
	if len(lel) != 40 {
		panic("DHT key with wrong length! WTF!")
	}
	return base32.StdEncoding.EncodeToString(lel)
}
