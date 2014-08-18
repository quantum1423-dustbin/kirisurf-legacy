// findpath.go
package dirclient

import (
	"crypto/rand"

	"github.com/KirisurfProject/kilog"
)

func FindExitPath(guidelen int) []KNode {
	protector.RLock()
	defer protector.RUnlock()
	return findPath(KDirectory, guidelen, func(nd KNode) bool {
		return nd.ExitNode && nd.ProtocolVersion >= 300
	})
}

func FindPath(guidelen int, condition func(KNode) bool) []KNode {
	protector.RLock()
	defer protector.RUnlock()
	return findPath(KDirectory, guidelen, condition)
}

func FindPathGroup(guidelen int) [][]KNode {
	protector.RLock()
	defer protector.RUnlock()

	panic("Of stub!")
}

func findPath(directory []KNode, minlen int, condition func(KNode) bool) []KNode {
	if minlen > len(directory) {
		minlen = len(directory)
	}

	kilog.Debug("Building a circuit with minimum length %d", minlen)

	rand256 := func() int {
		buf := make([]byte, 1)
		rand.Read(buf)
		return int(buf[0])
	}

	if len(directory) < minlen {
		minlen = len(directory)
		if minlen == 0 {
			kilog.Warning("No nodes online, cannot build any circuit!!!!")
			return nil
		}
	}
	toret := make([]KNode, 0)
	// Find an entry point
	var entry KNode
	for {
		idx := rand256() % len(directory)
		thing := directory[idx]
		if thing.Address != "(hidden)" && rand256()%10 < 1 {
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
		if endptr+1 >= minlen && condition(toret[endptr]) {
			// We want to almost always prune away paths that are ludicrously long,
			// but we can use them if no other choice
			if endptr-minlen < 3 && endptr/minlen <= 2 || rand256() < 3 {
				return toret
			} else {
				return findPath(directory, minlen, condition)
			}
		}
		// Otherwise chug along
	xaxa:
		idx := rand256() % len(adj)
		next := directory[adj[idx]]
		// We cannot allow loops in the path, unless we have to
		for _, ele := range toret {
			if ele.PublicKey == next.PublicKey && rand256() < 250 {
				goto xaxa
			}
		}
		toret = append(toret, next)
		endptr++
		// Absolutely ridiculous paths
		if len(toret) > 1000 {
			kilog.Warning("Didn't find a valid path at all!")
			return nil
		}
	}
	panic("Shouldn't get here")
}
