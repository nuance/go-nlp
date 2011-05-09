package features

import "fmt"
import "gnlp"
import "strings"

type Word string

func (w Word) Combine(o gnlp.Feature) gnlp.Feature {
	ow := o.(Word)

	return Word(fmt.Sprintf("%s %s", string(w), string(ow)))
}

func (w Word) Split() (gnlp.Feature, gnlp.Feature) {
	result := strings.Split(string(w), " ", 1)

	if len(result) == 1 {
		return Word(""), w
	}

	return Word(result[0]), Word(result[1])
}

func (w Word) String() string {
	return string(w)
}

