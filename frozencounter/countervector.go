package frozencounter

import "gnlp/minimizer"

// A countervector stores counters indexed by strings
type CounterVector struct {
	Keys *KeySet
	SubKeys *KeySet
	size int
	values vector
}

func NewCounterVector(counters map[string]*Counter) *CounterVector {
	var subks *KeySet = nil

	keys := []string{}
	for key, c := range counters {
		subks = c.Keys
		keys = append(keys, key)
	}

	ks := NewKeySet(keys, 0.0)
	size := len(subks.Keys)

	vals := make(vector, size * len(keys))
	for pos, key := range ks.Keys {
		copy(vals[pos:pos+size], counters[key].values)
	}

	return &CounterVector{Keys: ks, SubKeys: subks, size: size, values: vals}
}

// Copy the shape (but not the values) of cv
func (cv *CounterVector) Clone() *CounterVector {
	return &CounterVector{Keys: cv.Keys, SubKeys: cv.SubKeys, size: cv.size, values: make(vector, len(cv.values))}
}

func (cv *CounterVector) Extract() map[string]*Counter {
	result := make(map[string]*Counter)
	for _, key := range cv.Keys.Keys {
		result[key] = cv.Get(key)
	}

	return result
}

func (cv *CounterVector) Get(key string) *Counter {
	pos, ok := cv.Keys.Positions[key]
	if !ok {
		panic("Key missing")
	}

	return &Counter{Keys: cv.SubKeys, values: cv.values[pos*cv.size:(pos+1)*cv.size]}
}

func (cv *CounterVector) Set(key string, c Counter) {
	pos, ok := c.Keys.Positions[key]
	if !ok {
		panic("Key missing")
	}

	copy(c.values[pos*cv.size:(pos+1)*cv.size], c.values)
}

func (cv *CounterVector) check(o minimizer.Vector) {
	if cv.Keys != o.(*CounterVector).Keys || cv.SubKeys != o.(*CounterVector).SubKeys {
		panic("incompatible keysets")
	}
}

func (cv *CounterVector) Subtract(o minimizer.Vector) {
	cv.check(o)

	cv.AddScaled(-1.0, o.(*CounterVector))
}

func (cv *CounterVector) AddScaled(scale float64, o minimizer.Vector) {
	cv.check(o)

	cv.values.addScaled(scale, o.(*CounterVector).values)
}

func (cv *CounterVector) Negate() {
	cv.values.scale(-1.0)
}

func (cv *CounterVector) Scale(scale float64) {
	cv.values.scale(scale)
}

func (cv *CounterVector) Copy() minimizer.Vector {
	return &CounterVector{Keys: cv.Keys, SubKeys: cv.SubKeys, size: cv.size, values: cv.values.copy()}
}

func (cv *CounterVector) DotProduct(o minimizer.Vector) float64 {
	cv.check(o)

	return cv.values.dot(o.(*CounterVector).values)
}

