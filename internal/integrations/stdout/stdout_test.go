package stdout

import (
	"testing"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
)

func TestFormatReviewRequests(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		requests []ping.PingRequest
		expected string
	}{
		{
			name:     "No reviewers",
			requests: []ping.PingRequest{},
			expected: "No pending review requests.",
		},
		{
			name: "Only active reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-2 * time.Hour), IsTeam: true}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			expected: "Pinging: reviewer1 (1h ago, delay: 3600s), team1 (team) (2h ago, delay: 3600s)\n",
		},
		{
			name: "Only waiting reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-45 * time.Minute), IsTeam: true}, Enabled: true, Delay: 3600, ShouldPing: false},
			},
			expected: "Not pinging: reviewer1 (1h ago, delay: 3600s), status: waiting, team1 (team) (1h ago, delay: 3600s), status: waiting",
		},
		{
			name: "Only disabled reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)}, Enabled: false, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-3 * time.Hour), IsTeam: true}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			expected: "Not pinging: reviewer1 (2h ago, delay: 3600s), status: disabled, team1 (team) (3h ago, delay: 3600s), status: disabled",
		},
		{
			name: "Mixed reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-3 * time.Hour), IsTeam: true}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			expected: "Pinging: reviewer1 (2h ago, delay: 3600s)\nNot pinging: reviewer2 (1h ago, delay: 3600s), status: waiting, team1 (team) (3h ago, delay: 3600s), status: disabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatPingRequests(tc.requests)
			if result != tc.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tc.expected, result)
			}
		})
	}
}
