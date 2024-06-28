package gurps

import (
	"hash"
	"hash/fnv"

	"github.com/richardwilkes/toolbox/tid"
)

// Hashable is an object that can be hashed.
type Hashable interface {
	Hash(hash.Hash)
}

// Hash64 returns a 64-bit hash of the Hashable object.
func Hash64(in Hashable) uint64 {
	h := fnv.New64()
	in.Hash(h)
	return h.Sum64()
}

// NodesToHashesByID traverses the provided nodes and generates hashes.
func NodesToHashesByID[T NodeTypes](result map[tid.TID]uint64, data ...T) {
	Traverse(func(one T) bool {
		node := AsNode(one)
		id := node.ID()
		if _, exists := result[id]; !exists {
			result[id] = Hash64(node)
		}
		return false
	}, false, false, data...)
}
