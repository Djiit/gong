package stdout

import (
	"testing"
)

func TestFormatReviewers(t *testing.T) {
	result := formatReviewers([]string{"reviewer1", "reviewer2"})
	expected := "Awaiting reviews from: reviewer1, reviewer2"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
