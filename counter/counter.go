package counter

import "math"

type Counter struct {
	values map[string] float64
	// default value for missing items
	base float64
}

func New(base float64) *Counter {
	return &Counter{make(map[string] float64), base}
}

// Return a value for a key (falling back to the default)
func (c *Counter) Get(k string) float64 {
	v, ok := c.values[k]

	if ok {
		return v
	}
	return c.base
}

// Set a value for a key
func (c *Counter) Set(k string, v float64) {
	if v == c.Get(k) {
		return
	}

	c.values[k] = v
}

// Return a list of keys for this counter
func (c *Counter) Keys() []string {
	result := make([]string, len(c.values))

	for k, v := range c.values {
		// Don't track default values
		if v == c.base {
			continue
		}

		result = append(result, k)
	}

	return result
}

// Combine two sets of keys w/o duplicates
func mergeKeys(a, b []string) <-chan string {
	out := make(chan string)

	go func(out chan<- string) {
		seen := make(map[string] bool)

		for _, k := range a {
			out <- k
			seen[k] = true
		}

		for _, k := range b {
			if ! seen[k] {
				out <- k
			}
		}
	}(out)

	return out
}

// Apply an operation on two counters, returning new counter with keys
// defined by the keys function
func operate(a, b *Counter, op func (a, b float64) float64, keys func(a, b []string) <-chan string) *Counter {
	result := New(op(a.base, b.base))

	for k := range keys(a.Keys(), b.Keys()) {
		result.Set(k, op(a.Get(k), b.Get(k)))
	}

	return result
}

// Add a to b, returning a new counter
func Add(a, b *Counter) *Counter {
	return operate(a, b, func (a, b float64) float64 { return a + b }, mergeKeys)
}

// Subtract b from a, returning a new counter
func Subtract(a, b *Counter) *Counter {
	return operate(a, b, func (a, b float64) float64 { return a - b }, mergeKeys)
}

// Multiply a by b, returning a new counter
func Multiply(a, b *Counter) *Counter {
	return operate(a, b, func (a, b float64) float64 { return a * b }, mergeKeys)
}

// Divide a by b, returning a new counter
func Divide(a, b *Counter) *Counter {
	return operate(a, b, func (a, b float64) float64 { return a / b }, mergeKeys)
}

// Apply an operation on two counters, updating the first counter with keys
// defined by the keys function
func (a *Counter) operate(b *Counter, op func (a, b float64) float64, keys func(a, b []string) <-chan string) {
	a.base = op(a.base, b.base)

	for k := range keys(a.Keys(), b.Keys()) {
		a.Set(k, op(a.Get(k), b.Get(k)))
	}
}

// Add o to c
func (c *Counter) Add(o *Counter) {
	c.operate(o, func (a, b float64) float64 { return a + b }, mergeKeys)
}

// Subtract o from c
func (c *Counter) Subtract(o *Counter) {
	c.operate(o, func (a, b float64) float64 { return a - b }, mergeKeys)
}

// Multiply c by o
func (c *Counter) Multiply(o *Counter) {
	c.operate(o, func (a, b float64) float64 { return a * b }, mergeKeys)
}

// Divide c by o
func (c *Counter) Divide(o *Counter) {
	c.operate(o, func (a, b float64) float64 { return a / b }, mergeKeys)
}

// Apply a function to every value in the counter (including the
// default)
func (c *Counter) apply(op func (a float64) float64) {
	c.base = op(c.base)

	for k, v := range c.values {
		c.Set(k, op(v))
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

