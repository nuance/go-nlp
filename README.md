GNLP
====

A few structures for doing NLP analysis / experiments.

Basics
------

* counter.Counter

A map-like data structure for representing discrete probability
distributions. Contains an underlying map of event -> probability
along with a probability for all other events. Supports some
element-wise mathematical operations with other counter.Counter
objects.

```go
// Create a counter with 0 probability for unknown events (and with ""
// corresponding to the unknown event)
balls := counter.New(0.0)
	
// Add some observations
balls.Incr("blue")
balls.Incr("blue")
balls.Incr("red")

// Normalize into a discrete distribution
balls.Normalize()

// blue => 0.666666
balls.Get("blue")

// purple => 0.0
balls.Get("purple")

preference = counter.New(0.0)
preference.Set("red", 2.0)
preference.Set("blue", 1.0)
preference.Normalize()

expected_with_preference = counter.Multiply(balls, preference)
expected_with_preference.Normalize()

// blue => 0.5
expected_with_preference.Get("blue")
// red => 0.5
expected_with_preference.Get("red")

// You can also use log probabilities
balls.LogNormalize()
preferences.LogNormalize()

// And do in-place operations
balls.Add(preferences)

// Log-normalize expects counters with positive counts, so
// exponentiate-then-normalize
balls.Exp()
balls.LogNormalize()

// blue => -1 (== lg(0.5))
balls.Get("blue")
```

* frozencounter.Counter

Similar to counter.Counters, but with a fixed set of keys and no
default value. Represented under the hood as an array of doubles (with
order fixed according to the set of keys). Supports element-wise math
operations with other frozencounter.Counters that share the same set
of keys. Some mathematical operations are accelerated by the BLAS
library.
