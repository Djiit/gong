package githubclient

import (
	"testing"
	"time"
)

func TestFilterReviewRequestsByDelay(t *testing.T) {
	// Save original time.Now and replace with mock
	originalTimeNow := timeNow
	defer func() { timeNow = originalTimeNow }()

	now := time.Now()
	timeNow = func() time.Time { return now }

	testCases := []struct {
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
				{From: "team1", On: now.Add(-20 * time.Second), IsTeam: true},
			},
			expectedCount: 2, // All reviewers should be included with no delay
		},
		{
			name:  "negative delay",
			delay: -5,
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
				{From: "team1", On: now.Add(-20 * time.Second), IsTeam: true},
			},
			expectedCount: 2, // All reviewers should be included with negative delay
		},
		{
			name:  "with delay that filters some reviewers",
			delay: 15,
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false}, // 10s < 15s delay, should be filtered out
				{From: "team1", On: now.Add(-20 * time.Second), IsTeam: true},  // 20s > 15s delay, should be included
			},
			expectedCount: 1,
		},
		{
			name:  "with delay that filters all reviewers",
			delay: 30,
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
				{From: "team1", On: now.Add(-20 * time.Second), IsTeam: true},
			},
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filteredRequests := FilterReviewRequestsByDelay(tc.requests, tc.delay)
			if len(filteredRequests) != tc.expectedCount {
				t.Errorf("Expected %d filtered requests, got %d", tc.expectedCount, len(filteredRequests))
			}
		})
	}
}

func TestFilterReviewRequestsByRules(t *testing.T) {
	// Save original time.Now and replace with mock
	originalTimeNow := timeNow
	defer func() { timeNow = originalTimeNow }()

	now := time.Now()
	timeNow = func() time.Time { return now }

	testCases := []struct {
		name          string
		globalDelay   int
		rules         []Rule
		requests      []ReviewRequest
		expectedCount int
	}{
		{
			name:        "no rules, apply global delay",
			globalDelay: 15,
			rules:       []Rule{},
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
				{From: "team1", On: now.Add(-20 * time.Second), IsTeam: true},
			},
			expectedCount: 1, // Only the team1 exceeds the global delay
		},
		{
			name:        "rule matches user, overrides global delay",
			globalDelay: 30, // Would filter out all reviewers
			rules: []Rule{
				{MatchName: "user*", Delay: 5}, // Less strict delay for users
			},
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false}, // 10s > 5s rule delay
				{From: "team1", On: now.Add(-20 * time.Second), IsTeam: true},  // 20s < 30s global delay
			},
			expectedCount: 1, // Only user1 matches the rule and exceeds its delay
		},
		{
			name:        "multiple rules with different patterns",
			globalDelay: 30,
			rules: []Rule{
				{MatchName: "user*", Delay: 5},    // Matches user1
				{MatchName: "team*", Delay: 25},   // Matches team1 but with higher delay
				{MatchName: "org/*", Delay: 10},   // Doesn't match any reviewer
			},
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false}, // 10s > 5s delay from first rule
				{From: "team1", On: now.Add(-20 * time.Second), IsTeam: true},  // 20s < 25s delay from second rule
			},
			expectedCount: 1, // Only user1 exceeds its rule-specific delay
		},
		{
			name:        "rules with zero delay bypass filtering",
			globalDelay: 30,
			rules: []Rule{
				{MatchName: "user*", Delay: 0}, // Zero delay means include immediately
				{MatchName: "team*", Delay: 30},
			},
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
				{From: "team1", On: now.Add(-20 * time.Second), IsTeam: true},
			},
			expectedCount: 1, // user1 is included due to zero delay rule
		},
		{
			name:        "rules with negative delay bypass filtering",
			globalDelay: 30,
			rules: []Rule{
				{MatchName: "user*", Delay: -5}, // Negative delay means include immediately
			},
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
			},
			expectedCount: 1, // user1 is included due to negative delay rule
		},
		{
			name:        "exact match pattern",
			globalDelay: 30,
			rules: []Rule{
				{MatchName: "user1", Delay: 5}, // Exact match
				{MatchName: "user2", Delay: 5}, // No match
			},
			requests: []ReviewRequest{
				{From: "user1", On: now.Add(-10 * time.Second), IsTeam: false},
			},
			expectedCount: 1, // user1 matches exactly and exceeds delay
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filteredRequests := FilterReviewRequestsByRules(tc.requests, tc.globalDelay, tc.rules)
			if len(filteredRequests) != tc.expectedCount {
				t.Errorf("Expected %d filtered requests, got %d", tc.expectedCount, len(filteredRequests))
			}
		})
	}
}
