package main

import "fmt"
import F "gnlp/features"

func datum(class F.Word, s string) Datum {
	return Datum{class, []F.Word{F.Word(s)}}
}

func main() {
	training := []Datum{datum("name", "matt"), datum("name", "fred"), datum("name", "matt"), datum("pet", "matt")}

	nb := Train(training)
	class, prob := nb.Classify([]F.Word{F.Word("matt")})

	fmt.Printf("Guessed class %s w/ prob %.2f%%\n", class, prob*100)
}
