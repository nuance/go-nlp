package main

import "log"
import counter "gnlp/counter"
import frozencounter "gnlp/frozencounter"
import minimizer "gnlp/minimizer"

type MaxEnt struct {
	Weights map[string]*frozencounter.Counter
	Counts map[string]*frozencounter.Counter
	scorer *maxentWeights
}

type Datum struct {
	class    string
	features []string

	featureCounts *frozencounter.Counter
}

func tally(data []Datum) (counts map[string]*frozencounter.Counter, features *frozencounter.KeySet, labels []string) {
	rawCounts := map[string]*counter.Counter{}

	datumCounts := []*counter.Counter{}
	for _, datum := range data {
		if rawCounts[datum.class] == nil {
			rawCounts[datum.class] = counter.New(0.0)
		}
		c := counter.New(0.0)

		for _, f := range datum.features {
			rawCounts[datum.class].Incr(f)
			c.Incr(f)
		}

		datumCounts = append(datumCounts, c)
	}

	for idx, c := range frozencounter.FreezeMany(datumCounts) {
		data[idx].featureCounts = c
	}

	counts = frozencounter.FreezeMap(rawCounts)

	features = data[0].featureCounts.Keys
	for label, _ := range counts {
		labels = append(labels, label)
	}
	return
}

type maxentWeights struct {
	sigma float64
	data []Datum
	counts map[string]*frozencounter.Counter
	labels []string
	features *frozencounter.KeySet
}

// Calculate the label distribution of features given weights, storing the result in out
func (w *maxentWeights) labelDistribution(counts *frozencounter.Counter, weights map[string]*frozencounter.Counter) *frozencounter.Counter {
	out := counter.New(0.0)

	for label, featureWeights := range weights {
		out.Set(label, featureWeights.DotProduct(counts))
	}

	out.LogNormalize()
	return frozencounter.Freeze(out)
}

// given distribution for each datum, what's the expected count
func (w *maxentWeights) expectedCounts(labelDistribution []*frozencounter.Counter) (expectedCounts map[string]*frozencounter.Counter) {
	for _, label := range w.labels {
		expectedCounts[label] = frozencounter.New(w.features)
	}

	for idx, datum := range w.data {
		for _, label := range w.labels {
			counts := frozencounter.Multiply(datum.featureCounts, labelDistribution[idx])
			expectedCounts[label].Add(counts)
		}
	}

	return
}

func (w *maxentWeights) InitialWeights() (weights map[string]*frozencounter.Counter) {
	weights = map[string]*frozencounter.Counter{}
	for _, label := range w.labels {
		weights[label] = frozencounter.New(w.features)
	}
	return
}

func (w *maxentWeights) Gradient(weights map[string]*frozencounter.Counter) (value float64, gradient map[string]*frozencounter.Counter) {
	value = 0.0

	labelProbs := []*frozencounter.Counter{}
	for _, datum := range w.data {
		labelLogProbs := w.labelDistribution(datum.featureCounts, weights)
		value -= labelLogProbs.Get(datum.class)

		labelLogProbs.Exp()
		labelProbs = append(labelProbs, labelLogProbs)
	}

	gradient = w.expectedCounts(labelProbs)
	for label, featureCounts := range gradient {
		featureCounts.Subtract(w.counts[label])
	}

	// And penalize
	if w.sigma != 0.0 {
		penalty := 0.0

		for label, featureWeights := range weights {
			sqSums := featureWeights.Copy()
			sqSums.Apply(func(f *string, a float64) float64 { return a * a })
			penalty += sqSums.Sum()

			penalizedWeights := featureWeights.Copy()
			penalizedWeights.Scale(1 / (w.sigma * w.sigma))
			gradient[label].Add(penalizedWeights)
		}

		penalty /= 2 * w.sigma * w.sigma
		value += penalty
	}

	return
}

func (w *maxentWeights) Value(weights map[string]*frozencounter.Counter) (value float64) {
	value = 0.0

	for _, datum := range w.data {
		labelLogProbs := w.labelDistribution(datum.featureCounts, weights)
		value -= labelLogProbs.Get(datum.class)
	}

	// And penalize
	if w.sigma != 0.0 {
		penalty := 0.0

		for _, featureWeights := range weights {
			sqSums := featureWeights.Copy()
			sqSums.Apply(func(f *string, a float64) float64 { return a * a })
			penalty += sqSums.Sum()
		}

		penalty /= 2 * w.sigma * w.sigma
		value += penalty
	}

	return
}

func Train(data []Datum, l *log.Logger) *MaxEnt {
	l.Println("Building features")
	counts, features, labels := tally(data)

	weightFn := &maxentWeights{sigma: 0.01, data: data, counts: counts, features: features, labels: labels}
	l.Println("Minimizing")
	weights := minimizer.GradientDescent(minimizer.Standard, weightFn, l)

	return &MaxEnt{Counts: counts, Weights: weights, scorer: weightFn}
}

func (me *MaxEnt) Classify(features []string) (label string, score float64) {
	counts := frozencounter.New(me.scorer.features)
	for _, feature := range features {
		counts.Incr(feature)
	}

	logProbs := me.scorer.labelDistribution(counts, me.Weights)

	label, score = logProbs.ArgMax()
	return
}
