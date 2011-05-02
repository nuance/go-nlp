package frozencounter

import counter "gnlp/counter"
import crc     "hash/crc64"
import "fmt"
import "math"

type KeySet struct {
	Keys      []string
	Positions map[string]int
	Hash      uint64
	Base      float64
}

type Counter struct {
	Keys   *KeySet
	values vector
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

func New(ks *KeySet) *Counter {
	v := make(vector, len(ks.Keys))
	v.reset(ks.Base)

	return &Counter{ks, v}
}

// Freeze a counter, using a previously-generated keyset and
// index.
func FreezeWithKeySet(c *counter.Counter, ks *KeySet) *Counter {
	values := make([]float64, len(ks.Keys))

	for s, idx := range ks.Positions {
		values[idx] = c.Get(s)
	}

	return &Counter{ks, values}
}

// Convert a counter.Counter into a frozen counter, returning the new
// frozen counter and the index required to convert it back.
func Freeze(c *counter.Counter) *Counter {
	ks := NewKeySet(c.Keys(), c.Base)

	return FreezeWithKeySet(c, ks)
}

func mergeKeys(counters []*counter.Counter) []string {
	keys := make(map[string]bool)

	for _, c := range counters {
		for _, k := range c.Keys() {
			keys[k] = true
		}
	}

	r := make([]string, 0, len(keys))
	for k, _ := range keys {
		r = append(r, k)
	}

	return r
}

// Freeze multiple counters using the same index / keyset
func FreezeMany(counters []*counter.Counter) []*Counter {
	results := make([]*Counter, 0, len(counters))
	if len(counters) == 0 {
		return results
	}
	ks := NewKeySet(mergeKeys(counters), counters[0].Base)
	for _, c := range counters {
		results = append(results, FreezeWithKeySet(c, ks))
	}

	return results
}

func FreezeMap(counters map[string]*counter.Counter) map[string]*Counter {
	order := make([]string, 0, len(counters))
	for k, _ := range counters {
		order = append(order, k)
	}

	dists := make([]*counter.Counter, 0, len(counters))
	for _, dist := range counters {
		dists = append(dists, dist)
	}

	frozenList := FreezeMany(dists)

	frozen := make(map[string]*Counter)
	for idx, feature := range order {
		frozen[feature] = frozenList[idx]
	}

	return frozen
}

// Convert a frozen counter back into a counter.Counter.
func (c *Counter) Thaw(base float64) *counter.Counter {
	t := counter.New(base)

	for s, idx := range c.Keys.Positions {
		t.Set(s, c.values[idx])
	}

	return t
}

func (c *Counter) Copy() *Counter {
	return &Counter{c.Keys, c.values.copy()}
}

func (c *Counter) String() string {
	s := "FrozenCounter: {"

	for idx, key := range c.Keys.Keys {
		s += fmt.Sprintf("'%s': %f, ", key, c.values[idx])
	}

	s += "}"

	return s
}

func (c *Counter) ArgMax() (string, float64) {
	idx := c.values.argmax()

    return c.Keys.Keys[idx], c.values[idx]
}

// Add a to b, returning a new counter
func Add(a, b *Counter) *Counter {
	result := a.Copy()
	result.Add(b)
	return result
}

// Subtract b from a, returning a new counter
func Subtract(a, b *Counter) *Counter {
	result := a.Copy()
	result.Subtract(b)
	return result
}

// Multiply a by b, returning a new counter
func Multiply(a, b *Counter) *Counter {
	result := a.Copy()
	result.Multiply(b)
	return result
}

// Divide a by b, returning a new counter
func Divide(a, b *Counter) *Counter {
	result := a.Copy()
	result.Divide(b)
	return result
}

// Apply an operation on two counters, updating the first counter
func (a *Counter) operate(b *Counter, op func(a, b float64) float64) {
	if a.Keys != b.Keys {
		panic("Incoompatible frozen counters")
	}

	for idx, val := range a.values {
		a.values[idx] = op(val, b.values[idx])
	}
}

// Add o to c
func (c *Counter) Add(o *Counter) {
	if c.Keys != o.Keys {
		panic("Incoompatible frozen counters")
	}

	c.values.add(o.values)
}

// Subtract o from c
func (c *Counter) Subtract(o *Counter) {
	if c.Keys != o.Keys {
		panic("Incoompatible frozen counters")
	}

	c.values.subtract(o.values)
}

// Element-wise multiply c by o. Note that this is not blas-accelerated.
func (c *Counter) Multiply(o *Counter) {
	c.operate(o, func(a, b float64) float64 { return a * b })
}

// Element-wise divide c by o. Note that this is not blas-accelerated.
func (c *Counter) Divide(o *Counter) {
	c.operate(o, func(a, b float64) float64 { return a / b })
}

// Apply a function to every value in the counter
func (c *Counter) apply(op func(a float64) float64) {
	for idx, v := range c.values {
		c.values[idx] = op(v)
	}
}

// Log every value in the counter (including the default)
func (c *Counter) Log() {
	c.apply(math.Log)
}

// Exponentiate every value in the counter (including the default)
func (c *Counter) Exp() {
	c.apply(math.Exp)
}

// Normalize a counter s.t. the sum over values is now 1.0
func (c *Counter) Normalize() {
	sum := c.values.sum()
	c.values.scale(1.0 / sum)
}

// Special case of normalize - normalize a distribution and turn it
// into a log-distribution (performing the normalization after the
// xform to maintain precision)
func (c *Counter) LogNormalize() {
	sum := c.values.sum()
	logSum := math.Log(sum)

	c.apply(func(a float64) float64 { return math.Log(a) - logSum })
}
