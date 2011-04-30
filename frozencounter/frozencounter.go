package frozencounter

import counter "gnlp/counter"
import crc "hash/crc64"
import "math"

type KeySet struct {
	Keys []string
	Positions map[string] int
	Hash uint64
}

type Counter struct {
	keys *KeySet
	values []float64
}

// Build a key set of the keys + a crc64 of the keys (which we can
// efficiently compare). Also returns an index of string to position
func NewKeySet(keys []string) *KeySet {
	c := crc.New(crc.MakeTable(crc.ISO))
	index := make(map[string] int)

	for idx, s := range keys {
		index[s] = idx
		c.Write([]byte(s))
	}

	return &KeySet{Hash: c.Sum64(), Keys: keys, Positions: index}
}

func New(ks *KeySet) *Counter {
	return &Counter{ks, make([]float64, len(ks.Keys))}
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
	ks := NewKeySet(c.Keys())

	return FreezeWithKeySet(c, ks)
}

// Freeze multiple counters using the same index / keyset
func FreezeMany(counters []*counter.Counter) []*Counter {
	results := make([]*Counter, len(counters))
	if len(counters) == 0 {
		return results
	}

	ks := NewKeySet(counters[0].Keys())
	for _, c := range counters {
		results = append(results, FreezeWithKeySet(c, ks))
	}

	return results
}

// Convert a frozen counter back into a counter.Counter.
func (c *Counter) Thaw(base float64) *counter.Counter {
	t := counter.New(base)

	for s, idx := range c.keys.Positions {
		t.Set(s, c.values[idx])
	}

	return t
}

// Apply an operation on two counters, returning new counter with keys
// defined by the keys function
func operate(a, b *Counter, op func (a, b float64) float64) *Counter {
	result := New(a.keys)

	for idx, val := range a.values {
		result.values[idx] = op(val, b.values[idx])
	}

	return result
}

// Add a to b, returning a new counter
func Add(a, b *Counter) *Counter {
	return operate(a, b, func (a, b float64) float64 { return a + b })
}

// Subtract b from a, returning a new counter
func Subtract(a, b *Counter) *Counter {
	return operate(a, b, func (a, b float64) float64 { return a - b })
}

// Multiply a by b, returning a new counter
func Multiply(a, b *Counter) *Counter {
	return operate(a, b, func (a, b float64) float64 { return a * b })
}

// Divide a by b, returning a new counter
func Divide(a, b *Counter) *Counter {
	return operate(a, b, func (a, b float64) float64 { return a / b })
}

// Apply an operation on two counters, updating the first counter
func (a *Counter) operate(b *Counter, op func (a, b float64) float64) {
	if a.keys.Hash != b.keys.Hash {
		panic("Incoompatible frozen counters")
	}

	for idx, val := range a.values {
		a.values[idx] = op(val, b.values[idx])
	}
}

// Add o to c
func (c *Counter) Add(o *Counter) {
	c.operate(o, func (a, b float64) float64 { return a + b })
}

// Subtract o from c
func (c *Counter) Subtract(o *Counter) {
	c.operate(o, func (a, b float64) float64 { return a - b })
}

// Multiply c by o
func (c *Counter) Multiply(o *Counter) {
	c.operate(o, func (a, b float64) float64 { return a * b })
}

// Divide c by o
func (c *Counter) Divide(o *Counter) {
	c.operate(o, func (a, b float64) float64 { return a / b })
}

// Apply a function to every value in the counter
func (c *Counter) apply(op func (a float64) float64) {
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

// Reduce over the values in the counter (not including the default
// value)
func (c *Counter) reduce(base float64, op func (a, b float64) float64) float64 {
	val := base

	for _, v := range c.values {
		val = op(val, v)
	}

	return val
}

// Normalize a counter s.t. the sum over values is now 1.0
func (c *Counter) Normalize() {
	sum := c.reduce(0.0, func (a, b float64) float64 { return a + b })
	c.apply(func (a float64) float64 { return a / sum })
}

// Special case of normalize - normalize a distribution and turn it
// into a log-distribution (performing the normalization after the
// xform to maintain precision)
func (c *Counter) LogNormalize() {
	sum := c.reduce(0.0, func (a, b float64) float64 { return a + b })
	logSum := math.Log(sum)

	c.apply(func (a float64) float64 { return math.Log(a) - logSum })
}

