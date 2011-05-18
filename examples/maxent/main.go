package main

import "log"
import "fmt"
import "os"

func datum(class, s string) Datum {
	return Datum{class: class, features: []string{s}}
}

func main() {
	training := []Datum{datum("name", "matt"), datum("name", "fred"), datum("name", "matt"), datum("pet", "matt")}

	maxent := Train(training, log.New(os.Stderr, "[Maxent] ", log.LstdFlags))

	class, prob := maxent.Classify([]string{"matt"})
	fmt.Printf("Guessed class %s w/ prob %.2f%%\n", class, prob*100)
}
