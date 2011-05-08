package gnlp

type Feature interface {
	// combine another feature with this feature. For unigrams, this
	// means produce the bigram from this and f, like this: "dog", f: "hot"
	// return "hot dog"
	Combine(f Feature) Feature
	// Split this feature into a new feature and it's history. For a
	// trigram feature "good hot dog", return "good" and "hot dog"
	Split() (Feature, Feature)
	String() string
}

