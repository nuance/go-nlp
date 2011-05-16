package features

import "fmt"
import "strings"

func WordCombine(a, b string) string {

	return fmt.Sprintf("%s %s", a, b)
}

func WordSplit(w string) (string, string) {
	result := strings.Split(w, " ", 1)

	if len(result) == 1 {
		return "", w
	}

	return result[0], result[1]
}

