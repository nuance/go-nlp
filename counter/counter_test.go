package counter

import "testing"

func TestBasics(t *testing.T) {
	// Create a counter with 0 probability for unknown events
	balls := New(0.0)

	// Add some observations
	balls.Incr("blue")
	balls.Incr("blue")
	balls.Incr("red")

	// Normalize into a discrete distribution
	balls.Normalize()

	// blue => 0.66666666
	if balls.Get("blue") != 2.0 / 3.0 {
		t.Error("Blue doesn't have 2/3rds probability")
		t.FailNow()
	}

	// purple => 0.0
	if balls.Get("purple") != 0.0 {
		t.Error("Purple doesn't have 0 probability")
	}

	preference := New(0.0)
	preference.Set("red", 2.0)
	preference.Set("blue", 1.0)
	preference.Normalize()

	expected_with_preference := Multiply(balls, preference)
	expected_with_preference.Normalize()

	// blue => 0.5
	if expected_with_preference.Get("blue") != 0.5 {
		t.Error("Blue doesn't have 0.5 probability")
	}
	// red => 0.5
	if expected_with_preference.Get("red") != 0.5 {
		t.Error("Red doesn't have 0.5 probability")
	}
}
