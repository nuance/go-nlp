package smooth

import "gnlp"

func LaPlace(c gnlp.Counter, alpha float64) {
	c.Apply(func(s *string, w float64) float64 {
		return w + alpha
	})

	c.Normalize()
}

func GoodTuring(c gnlp.Counter, estimator func(key *string, count float64) float64) {
	c.Apply(func(key *string, w float64) float64 {
		return (w + 1) * estimator(key, w+1) / estimator(key, w)
	})

	c.Normalize()
}

// Good turing with a linear-combination fallback estimate
func JelinekMercer(counts, fallbackCounts gnlp.Counter, fallbackWeight func(key *string) float64, split func(key *string) (string, string)) {
	counts.Apply(func(key *string, w float64) float64 {
		_, smaller := split(key)
		weight := fallbackWeight(key)

		return (1-weight)*counts.Get(*key) + weight*fallbackCounts.Get(smaller)
	})

	counts.Normalize()
}
