package githubclient

import (
	"testing"
	"time"
)

func TestFilterReviewRequestsByDelay(t *testing.T) {
	// Create a fixed reference time for consistent testing
	now := time.Now()

	// Setup test cases
	tests := []struct {
		name          string
		delay         int
		requests      []ReviewRequest
		expectedCount int
	}{
		{
			name:  "no delay",
			delay: 0,
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
				{From: "user2", On: now.Add(-20 * time.Second), IsTeam: false},
			},
			expectedCount: 2, // All reviewers should be included with no delay
		},
		{
			name:  "negative delay",
			delay: -5,
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
				{From: "user2", On: now.Add(-20 * time.Second), IsTeam: false},
			},
			expectedCount: 2, // All reviewers should be included with negative delay
		},
		{
			name:  "with delay that filters some reviewers",
			delay: 15,
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false}, // Too recent, should be filtered
				{From: "user2", On: now.Add(-20 * time.Second), IsTeam: false}, // Old enough, should be included
			},
			expectedCount: 1, // Only user2 should be included
		},
		{
			name:  "with delay that filters all reviewers",
			delay: 30,
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
				{From: "user2", On: now.Add(-20 * time.Second), IsTeam: false},
			},
			expectedCount: 0, // No reviewers should be included
		},
		{
			name:          "empty request list",
			delay:         10,
			requests:      []ReviewRequest{},
			expectedCount: 0, // Empty list should remain empty
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Use a function that wraps time.Now() to make tests deterministic
			originalTimeNow := timeNow
			defer func() { timeNow = originalTimeNow }()

			// Mock the time.Now function to return our fixed time
			timeNow = func() time.Time {
				return now
			}

			// Call the function being tested
			filtered := FilterReviewRequestsByDelay(tc.requests, tc.delay)

			// Verify the results
			if len(filtered) != tc.expectedCount {
				t.Errorf("Expected %d reviewers after delay filter, got %d", tc.expectedCount, len(filtered))
			}

			// Check that the right reviewers were filtered
			if tc.expectedCount > 0 && tc.delay > 0 {
				for _, req := range filtered {
					if now.Sub(req.On).Seconds() < float64(tc.delay) {
						t.Errorf("Reviewer %s should have been filtered out (delay: %d, time since request: %v)",
							req.From, tc.delay, now.Sub(req.On))
					}
				}
			}
		})
	}
}
