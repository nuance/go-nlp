package frozencounter

import "gnlp"
import counter "gnlp/counter"
import "fmt"
import "math"

type Counter struct {
	Keys   *KeySet
	values vector
}

func New(ks *KeySet) *Counter {
	v := make(vector, len(ks.Keys))
	v.reset(ks.Base)

	return &Counter{ks, v}
}

func (c *Counter) Reset(v float64) {
	c.values.reset(v)
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

func (c *Counter) Get(f string) float64 {
	idx, ok := c.Keys.Positions[f]

	if !ok {
		return c.Keys.Base
	}

	return c.values[idx]
}

func (c *Counter) Set(f string, val float64) {
	idx, ok := c.Keys.Positions[f]

	if !ok {
		panic("Feature not found in frozen counter")
	}

	c.values[idx] = val
}

func (c *Counter) Incr(f string) {
	idx, ok := c.Keys.Positions[f]

	if !ok {
		panic("Feature not found in frozen counter")
	}

	c.values[idx] += 1
}

// Convert a frozen counter back into a counter.Counter.
func (c *Counter) Thaw() *counter.Counter {
	t := counter.New(c.Keys.Base)

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

func (c *Counter) check(o *Counter) {
	if c.Keys != o.Keys {
		fmt.Println(c)
		fmt.Println(o)
		panic("incompatible keysets")
	}
}

// Add a to b, returning a new counter
func Add(a, b *Counter) *Counter {
	a.check(b)

	result := a.Copy()
	result.Add(b)
	return result
}

// Subtract b from a, returning a new counter
func Subtract(a, b *Counter) *Counter {
	a.check(b)

	result := a.Copy()
	result.Subtract(b)
	return result
}

// Multiply a by b, returning a new counter
func Multiply(a, b *Counter) *Counter {
	a.check(b)

	result := a.Copy()
	result.Multiply(b)
	return result
}

// Divide a by b, returning a new counter
func Divide(a, b *Counter) *Counter {
	a.check(b)

	result := a.Copy()
	result.Divide(b)
	return result
}

func Dot(a, b *Counter) float64 {
	a.check(b)

	return a.values.dot(b.values)
}

// Apply an operation on two counters, updating the first counter
func (c *Counter) operate(o *Counter, op func(a, b float64) float64) {
	c.check(o)

	for idx, val := range c.values {
		c.values[idx] = op(val, o.values[idx])
	}
}

// Add o to c
func (c *Counter) Add(o *Counter) {
	c.check(o)

	c.values.add(o.values)
}

// Add o to c
func (c *Counter) AddScaled(scale float64, o *Counter) {
	c.check(o)

	c.values.addScaled(scale, o.values)
}

// Subtract o from c
func (c *Counter) Subtract(o *Counter) {
	c.check(o)

	c.values.subtract(o.values)
}

// Element-wise multiply c by o. Note that this is not blas-accelerated.
func (c *Counter) Multiply(o *Counter) {
	c.check(o)

	c.operate(o, func(a, b float64) float64 { return a * b })
}

func (c *Counter) Scale(val float64) {
	c.values.scale(val)
}

// Element-wise divide c by o. Note that this is not blas-accelerated.
func (c *Counter) Divide(o *Counter) {
	c.check(o)

	c.operate(o, func(a, b float64) float64 { return a / b })
}

// Compute the dot product of c & o.
func (c *Counter) DotProduct(o *Counter) float64 {
	c.check(o)

	return c.values.dot(o.values)
}

// Apply a function to every value in the counter
func (c *Counter) Apply(op func(f *string, a float64) float64) {
	for idx, v := range c.values {
		c.values[idx] = op(&c.Keys.Keys[idx], v)
	}
}

// Log every value in the counter (including the default)
func (c *Counter) Log() {
	c.Apply(func(f *string, a float64) float64 { return math.Log(a) })
}

// Exponentiate every value in the counter (including the default)
func (c *Counter) Exp() {
	c.Apply(func(f *string, a float64) float64 { return math.Exp(a) })
}

func (c *Counter) Sum() float64 {
	return c.values.sum()
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

	c.Apply(func(f *string, a float64) float64 { return math.Log(a) - logSum })
}

var _ gnlp.Counter = new(Counter)
