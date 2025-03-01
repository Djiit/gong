package stdout

import (
	"testing"
	"time"

	"github.com/Djiit/pingrequest/internal/githubclient"
)

func TestFormatReviewRequests(t *testing.T) {
	now := time.Now()
	result := formatReviewRequests([]githubclient.ReviewRequest{
		{From: "reviewer1", On: now.Add(-1 * time.Hour)},
		{From: "reviewer2", On: now.Add(-1 * time.Hour), IsTeam: true},
	})
	expected := "Awaiting reviews from: reviewer1 (1h ago), reviewer2 (team) (1h ago)"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
