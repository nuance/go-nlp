package main

import "gnlp"

import counter "gnlp/counter"
import F "gnlp/features"
import frozencounter "gnlp/frozencounter"

type NaiveBayes struct {
	FeatureLogDistributions map[gnlp.Feature]*frozencounter.Counter
	ClassLogPrior           *frozencounter.Counter
}

type Datum struct {
	class    F.Word
	features []F.Word
}

func Train(data []Datum) *NaiveBayes {
	class := counter.New(0.0)
	features := make(map[gnlp.Feature]*counter.Counter)

	for _, datum := range data {
		class.Incr(F.Word(datum.class))
		for _, f := range datum.features {
			dist, ok := features[f]

			if !ok {
				dist = counter.New(0.0)
				features[f] = dist
			}

			dist.Incr(F.Word(datum.class))
		}
	}

	class.LogNormalize()
	for _, dist := range features {
		dist.LogNormalize()
	}

	frozenFeatures := frozencounter.FreezeMap(features)

	var keyset *frozencounter.KeySet
	for _, dist := range frozenFeatures {
		keyset = dist.Keys
	}

	frozenClass := frozencounter.FreezeWithKeySet(class, keyset)

	return &NaiveBayes{FeatureLogDistributions: frozenFeatures, ClassLogPrior: frozenClass}
}

func (nb *NaiveBayes) Classify(features []F.Word) (F.Word, float64) {
	score := nb.ClassLogPrior.Copy()

	for _, f := range features {
		score.Add(nb.FeatureLogDistributions[f])
	}

	score.Exp()
	score.Normalize()

	c, probability := score.ArgMax()
	return c.(F.Word), probability
}
