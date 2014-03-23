// findpath.go
package dirclient

import "math/rand"

func FindPath(minlen int) []KNode {
	protector.RLock()
	defer protector.RUnlock()
	toret := make([]KNode, 0)
	// Find an entry point
	var entry KNode
	for {
		idx := rand.Int() % len(KDirectory)
		thing := KDirectory[idx]
		if thing.Address != "(hidden)" {
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
		if endptr+1 >= minlen && toret[endptr].ExitNode {
			return toret
		}
		// Otherwise chug along
		idx := rand.Int() % len(adj)
		next := KDirectory[adj[idx]]
		toret = append(toret, next)
		endptr++
	}
	panic("Shouldn't get here")
}
