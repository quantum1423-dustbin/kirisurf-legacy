// findpath.go
package dirclient

import (
	"crypto/rand"

	"github.com/KirisurfProject/kilog"
)

func FindPathGroup(guidelen int) [][]KNode {
	protector.RLock()
	defer protector.RUnlock()

	// We first find one path and obtain the endpoint
	protopath := FindPath(KDirectory, guidelen)
	endpoint := protopath[len(protopath)-1]

	// Make a copy of the directory
	DirCopy := make([]KNode, len(KDirectory))
	copy(DirCopy, KDirectory)

	// Unmark all nodes except endpoint as exit
	for i := 0; i < len(DirCopy); i++ {
		if DirCopy[i].Address != endpoint.Address {
			DirCopy[i].ExitNode = false
		}
	}

	// Return 16 random paths
	toret := make([][]KNode, 16)
	for i := 0; i < 16; i++ {
		toret[i] = FindPath(DirCopy, guidelen)
	}

	return toret
}

func FindPath(directory []KNode, minlen int) []KNode {
	if minlen > len(directory) {
		minlen = len(directory)
	}

	kilog.Debug("minlen = %d", minlen)

	rand256 := func() int {
		buf := make([]byte, 1)
		rand.Read(buf)
		return int(buf[0])
	}

	if len(directory) < minlen {
		minlen = len(directory)
		if minlen == 0 {
			panic("No nodes online, cannot build any circuit!!!!")
		}
	}
	toret := make([]KNode, 0)
	// Find an entry point
	var entry KNode
	for {
		idx := rand256() % len(directory)
		thing := directory[idx]
		if thing.Address != "(hidden)" && rand256()%10 < 1 && !thing.ExitNode {
			entry = thing
			break
		}
	}
	// Push the entry onto the slice
	toret = append(toret, entry)
	//history := make(map[int]bool)
	endptr := 0
	for {
		adj := toret[endptr].Adjacents
		// If already at the end, return
		if endptr+1 >= minlen && toret[endptr].ExitNode && toret[endptr].ProtocolVersion >= 300 {
			kilog.Debug("%v", toret)
			return toret
		}
		// Otherwise chug along
		idx := rand256() % len(adj)
		next := directory[adj[idx]]
		toret = append(toret, next)
		endptr++
	}
	panic("Shouldn't get here")
}
