package actions

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
)

func TestProcessRequests(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name             string
		requests         []ping.PingRequest
		expectedEnabled  []string
		expectedDisabled []string
	}{
		{
			name:             "No reviewers",
			requests:         []ping.PingRequest{},
			expectedEnabled:  []string{},
			expectedDisabled: []string{},
		},
		{
			name: "Only enabled reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-2 * time.Hour), IsTeam: true}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			expectedEnabled:  []string{"reviewer1", "team1 (team)"},
			expectedDisabled: []string{},
		},
		{
			name: "Only waiting reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-45 * time.Minute), IsTeam: true}, Enabled: true, Delay: 3600, ShouldPing: false},
			},
			expectedEnabled: []string{},
			expectedDisabled: []string{
				"reviewer1 (1h ago, status: waiting)",
				"team1 (team) (1h ago, status: waiting)",
			},
		},
		{
			name: "Only disabled reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)}, Enabled: false, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-3 * time.Hour), IsTeam: true}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			expectedEnabled: []string{},
			expectedDisabled: []string{
				"reviewer1 (2h ago, status: disabled)",
				"team1 (team) (3h ago, status: disabled)",
			},
		},
		{
			name: "Mixed reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-3 * time.Hour), IsTeam: true}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			expectedEnabled: []string{"reviewer1"},
			expectedDisabled: []string{
				"reviewer2 (1h ago, status: waiting)",
				"team1 (team) (3h ago, status: disabled)",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enabled, disabled := processRequests(tc.requests)

			// Check enabled reviewers
			if len(enabled) != len(tc.expectedEnabled) {
				t.Errorf("expected %d enabled reviewers, got %d", len(tc.expectedEnabled), len(enabled))
			}
			for i, reviewer := range enabled {
				if i < len(tc.expectedEnabled) && reviewer != tc.expectedEnabled[i] {
					t.Errorf("expected enabled reviewer '%s', got '%s'", tc.expectedEnabled[i], reviewer)
				}
			}

			// Check disabled reviewers
			if len(disabled) != len(tc.expectedDisabled) {
				t.Errorf("expected %d disabled reviewers, got %d", len(tc.expectedDisabled), len(disabled))
			}
			for i, reviewer := range disabled {
				if i < len(tc.expectedDisabled) && reviewer != tc.expectedDisabled[i] {
					t.Errorf("expected disabled reviewer '%s', got '%s'", tc.expectedDisabled[i], reviewer)
				}
			}
		})
	}
}

func TestFormatOutput(t *testing.T) {
	testCases := []struct {
		name              string
		enabledReviewers  []string
		disabledReviewers []string
		expected          string
	}{
		{
			name:              "No reviewers",
			enabledReviewers:  []string{},
			disabledReviewers: []string{},
			expected:          "No pending review requests.",
		},
		{
			name:              "Only enabled reviewers",
			enabledReviewers:  []string{"reviewer1", "team1 (team)"},
			disabledReviewers: []string{},
			expected:          "Enabled reviewers (2): reviewer1, team1 (team)\n",
		},
		{
			name:              "Only disabled reviewers",
			enabledReviewers:  []string{},
			disabledReviewers: []string{"reviewer1 (2h ago, status: disabled)", "team1 (team) (3h ago, status: disabled)"},
			expected:          "Disabled/waiting reviewers (2): reviewer1 (2h ago, status: disabled), team1 (team) (3h ago, status: disabled)",
		},
		{
			name:              "Mixed reviewers",
			enabledReviewers:  []string{"reviewer1"},
			disabledReviewers: []string{"reviewer2 (1h ago, status: waiting)", "team1 (team) (3h ago, status: disabled)"},
			expected:          "Enabled reviewers (1): reviewer1\nDisabled/waiting reviewers (2): reviewer2 (1h ago, status: waiting), team1 (team) (3h ago, status: disabled)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatOutput(tc.enabledReviewers, tc.disabledReviewers)
			if result != tc.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tc.expected, result)
			}
		})
	}
}

func TestWriteToGitHubOutput(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "actions-test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Fatalf("Error removing temp directory: %v", err)
		}
	}()

	outputFile := filepath.Join(tempDir, "github-output")

	testCases := []struct {
		name              string
		enabledReviewers  []string
		disabledReviewers []string
		expectedContains  []string
	}{
		{
			name:              "No reviewers",
			enabledReviewers:  []string{},
			disabledReviewers: []string{},
			expectedContains: []string{
				"reviewers=",
				"reviewersCount=0",
				"reviewersDetails<<EOF",
				"EOF",
			},
		},
		{
			name:              "Only enabled reviewers",
			enabledReviewers:  []string{"reviewer1", "team1 (team)"},
			disabledReviewers: []string{},
			expectedContains: []string{
				"reviewers=reviewer1,team1 (team)",
				"reviewersCount=2",
				"reviewersDetails<<EOF",
				"reviewer1 (status: enabled)",
				"team1 (team) (status: enabled)",
				"EOF",
			},
		},
		{
			name:              "With disabled reviewers",
			enabledReviewers:  []string{"reviewer1"},
			disabledReviewers: []string{"reviewer2 (1h ago, status: waiting)", "team1 (team) (3h ago, status: disabled)"},
			expectedContains: []string{
				"reviewers=reviewer1",
				"reviewersCount=1",
				"reviewersDetails<<EOF",
				"reviewer1 (status: enabled)",
				"reviewer2 (1h ago, status: waiting)",
				"team1 (team) (3h ago, status: disabled)",
				"EOF",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset file for each test
			err := os.WriteFile(outputFile, []byte{}, 0644)
			if err != nil {
				t.Fatalf("failed to reset output file: %v", err)
			}

			err = writeToGitHubOutput(outputFile, tc.enabledReviewers, tc.disabledReviewers)
			if err != nil {
				t.Fatalf("failed to write to GitHub output: %v", err)
			}

			// Read the file content
			content, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}

			fileContent := string(content)
			for _, expectedLine := range tc.expectedContains {
				if !strings.Contains(fileContent, expectedLine) {
					t.Errorf("expected output to contain '%s', but it doesn't: %s", expectedLine, fileContent)
				}
			}
		})
	}
}

func TestWriteToGitHubEnv(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "actions-test")
	if err != nil {
		t.Fatalf("failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Fatalf("Error removing temp directory: %v", err)
		}
	}()

	envFile := filepath.Join(tempDir, "github-env")

	testCases := []struct {
		name              string
		enabledReviewers  []string
		disabledReviewers []string
		expectedContains  []string
	}{
		{
			name:              "No reviewers",
			enabledReviewers:  []string{},
			disabledReviewers: []string{},
			expectedContains: []string{
				"GONG_REVIEWERS=",
				"GONG_REVIEWERS_COUNT=0",
				"GONG_REVIEWERS_DETAILS<<EOF",
				"EOF",
			},
		},
		{
			name:              "Only enabled reviewers",
			enabledReviewers:  []string{"reviewer1", "team1 (team)"},
			disabledReviewers: []string{},
			expectedContains: []string{
				"GONG_REVIEWERS=reviewer1,team1 (team)",
				"GONG_REVIEWERS_COUNT=2",
				"GONG_REVIEWERS_DETAILS<<EOF",
				"reviewer1 (status: enabled)",
				"team1 (team) (status: enabled)",
				"EOF",
			},
		},
		{
			name:              "With disabled reviewers",
			enabledReviewers:  []string{"reviewer1"},
			disabledReviewers: []string{"reviewer2 (1h ago, status: waiting)", "team1 (team) (3h ago, status: disabled)"},
			expectedContains: []string{
				"GONG_REVIEWERS=reviewer1",
				"GONG_REVIEWERS_COUNT=1",
				"GONG_REVIEWERS_DETAILS<<EOF",
				"reviewer1 (status: enabled)",
				"reviewer2 (1h ago, status: waiting)",
				"team1 (team) (3h ago, status: disabled)",
				"EOF",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset file for each test
			err := os.WriteFile(envFile, []byte{}, 0644)
			if err != nil {
				t.Fatalf("failed to reset env file: %v", err)
			}

			err = writeToGitHubEnv(envFile, tc.enabledReviewers, tc.disabledReviewers)
			if err != nil {
				t.Fatalf("failed to write to GitHub env: %v", err)
			}

			// Read the file content
			content, err := os.ReadFile(envFile)
			if err != nil {
				t.Fatalf("failed to read env file: %v", err)
			}

			fileContent := string(content)
			for _, expectedLine := range tc.expectedContains {
				if !strings.Contains(fileContent, expectedLine) {
					t.Errorf("expected env to contain '%s', but it doesn't: %s", expectedLine, fileContent)
				}
			}
		})
	}
}
