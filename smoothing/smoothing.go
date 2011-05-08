package smooth

import "gnlp"

func LaPlace(c gnlp.Counter, alpha float64) {
	c.Apply(func(s *gnlp.Feature, w float64) float64 {
		return w + alpha
	})

	c.Normalize()
}

func GoodTuring(c gnlp.Counter, estimator func(key *gnlp.Feature, count float64) float64) {
	c.Apply(func(key *gnlp.Feature, w float64) float64 {
		return (w + 1) * estimator(key, w+1) / estimator(key, w)
	})

	c.Normalize()
}

// Good turing with a linear-combination fallback estimate
func JelinekMercer(counts, fallbackCounts gnlp.Counter, fallbackWeight func(key *gnlp.Feature) float64) {
	counts.Apply(func(key *gnlp.Feature, w float64) float64 {
		_, smaller := (*key).Split()
		weight := fallbackWeight(key)

		return (1-weight)*counts.Get(*key) + weight*fallbackCounts.Get(smaller)
	})

	counts.Normalize()
}

func Katz(counts, fallbackCounts gnlp.Counter, reliableCutoff float64) {
	
}
