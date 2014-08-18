package kiridht

import (
	"encoding/binary"
	"kirisurf/ll/dirclient"
	"kirisurf/ll/kiss"
)

func GetDHTKey(nd dirclient.KNode) uint64 {
	return binary.LittleEndian.Uint64(kiss.KeyedHash([]byte(nd.PublicKey),
		[]byte("Kirisurf DHT magic thingy")))
}

func PickClosest(target uint64, candidates []uint64) uint64 {
	closest := candidates[0]
	for i := 0; i < len(candidates); i++ {
		if target-candidates[i] < target-closest {
			closest = target
		}
	}
	return closest
}
