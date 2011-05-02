package main

import "fmt"

func datum(class Class, s string) Datum {
	return Datum{class, []string{s}}
}

func main() {
	training := []Datum{datum("name", "matt"), datum("name", "fred"), datum("name", "matt"), datum("pet", "matt")}

	nb := Train(training)
	class, prob := nb.Classify([]string{"matt"})

	fmt.Printf("Guessed class %s w/ prob %.2f%%\n", class, prob * 100)
}
