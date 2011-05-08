package counter

import "gnlp"
import "fmt"
import "strings"
import "testing"

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

func TestBasics(t *testing.T) {
	// Create a counter with 0 probability for unknown events
	balls := New(0.0)

	// Add some observations
	balls.Incr(Word("blue"))
	balls.Incr(Word("blue"))
	balls.Incr(Word("red"))

	// Normalize into a discrete distribution
	balls.Normalize()

	// blue => 0.66666666
	if balls.Get(Word("blue")) != 2.0/3.0 {
		t.Error("Blue doesn't have 2/3rds probability")
		t.FailNow()
	}

	// purple => 0.0
	if balls.Get(Word("purple")) != 0.0 {
		t.Error("Purple doesn't have 0 probability")
	}

	preference := New(0.0)
	preference.Set(Word("red"), 2.0)
	preference.Set(Word("blue"), 1.0)
	preference.Normalize()

	expected_with_preference := Multiply(balls, preference)
	expected_with_preference.Normalize()

	// blue => 0.5
	if expected_with_preference.Get(Word("blue")) != 0.5 {
		t.Error("Blue doesn't have 0.5 probability")
	}
	// red => 0.5
	if expected_with_preference.Get(Word("red")) != 0.5 {
		t.Error("Red doesn't have 0.5 probability")
	}
}
