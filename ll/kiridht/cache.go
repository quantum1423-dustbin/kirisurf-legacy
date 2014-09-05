package kiridht

import "sync"

var cache = make(map[uint64][]byte)
var clock sync.RWMutex
var usage uint64
var USAGELIM = uint64(1024 * 1024 * 1024)

func cacheEvictOne() {
	clock.Lock()
	defer clock.Unlock()
	var key uint64
	for k, _ := range cache {
		key = k
		break
	}
	usage -= uint64(len(cache[key]))
	delete(cache, key)
}

func cacheAdd(key uint64, val []byte) {
	clock.Lock()
	defer clock.Unlock()
	for usage >= USAGELIM {
		cacheEvictOne()
	}
	cache[key] = val
	usage += uint64(len(val))
}

func cacheIdx(key uint64) []byte {
	clock.RLock()
	defer clock.RUnlock()
	return cache[key]
}
