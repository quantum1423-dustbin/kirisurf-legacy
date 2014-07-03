package intercom

import (
	"crypto/rand"
	"sync"
	"unsafe"
)

type ProbDistro map[int]int

var _pdlock sync.Mutex

func MakeProbDistro() ProbDistro {
	// uniform chance of anything from 32 to 1024 bytes
	toret := make(map[int]int)
	for i := 32; i <= 1056; i++ {
		xaxa := make([]byte, 1)
		rand.Read(xaxa)
		toret[i] = int(xaxa[0]) % 16
	}
	for i := 0; i < 10000; i++ {
		ProbDistro(toret).Juggle()
	}
	return toret
}

func (thing ProbDistro) Juggle() {
	_pdlock.Lock()
	defer _pdlock.Unlock()
start:

	randthing := make([]byte, 2)
	rand.Read(randthing)
	randidx1 := (int(*(*uint16)(unsafe.Pointer(&randthing[0]))) % 1024) + 32
	rand.Read(randthing)
	randidx2 := (int(*(*uint16)(unsafe.Pointer(&randthing[0]))) % 1024) + 32

	// Try to move randidx1 to randidx2.
	if thing[randidx1] == 0 {
		goto start
	}

	thing[randidx1]--
	thing[randidx2]++
}

func (thing ProbDistro) Draw() int {
	_pdlock.Lock()
	defer _pdlock.Unlock()

	xaxa := make([]int, 0)
	for i := 32; i <= 1056; i++ {
		if thing[i] != 0 {
			bloo := make([]int, thing[i])
			for j := 0; j < len(bloo); j++ {
				bloo[j] = i
			}
			xaxa = append(xaxa, bloo...)
		}
	}

	randthing := make([]byte, 4)
	rand.Read(randthing)
	randidx := (*(*uint32)(unsafe.Pointer(&randthing[0]))) % uint32(len(xaxa))
	return xaxa[randidx]
}
