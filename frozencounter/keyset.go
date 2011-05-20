package frozencounter

import crc "hash/crc64"

type KeySet struct {
	Keys      []string
	Positions map[string]int
	Hash      uint64
	Base      float64
}

var keySetCache map[uint64][]*KeySet

func init() {
	// Build the interning cache
	keySetCache = make(map[uint64][]*KeySet)
}

func internKeySet(ks *KeySet) *KeySet {
	possibles, ok := keySetCache[ks.Hash]

	// We found something with this hash
	if ok {
		// Look for it
		for _, possible := range possibles {
			// If the keys match up, this is it - return possible as
			// the canonical instance
			if possible.Base != ks.Base {
				continue
			}

			if len(possible.Keys) != len(ks.Keys) {
				continue
			}

			for idx, v := range possible.Keys {
				if ks.Keys[idx] != v {
					continue
				}
			}

			return possible
		}
	} else {
		// This bucket is empty, initialize it to a list
		keySetCache[ks.Hash] = make([]*KeySet, 0, 1)
	}

	// Put this keyset in the bucket as the canonical instance
	keySetCache[ks.Hash] = append(keySetCache[ks.Hash], ks)

	return ks

}

// Build a key set of the keys + a crc64 of the keys (which we can
// efficiently compare). Also returns an index of string to position
func NewKeySet(keys []string, base float64) *KeySet {
	c := crc.New(crc.MakeTable(crc.ISO))
	index := make(map[string]int)

	for idx, s := range keys {
		index[s] = idx
		c.Write([]byte(s))
	}

	return internKeySet(&KeySet{Hash: c.Sum64(), Keys: keys, Positions: index, Base: base})
}

