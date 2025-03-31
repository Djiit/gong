package comment

import (
	"testing"
	"time"

	"github.com/Djiit/gong/internal/format"
	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
)

func TestFormatWithTemplate(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name     string
		requests []ping.PingRequest
		template string
		expected string
	}{
		{
			name:     "No reviewers with default template",
			requests: []ping.PingRequest{},
			template: DefaultTemplate,
			expected: "No pending review requests.\n<!-- gong -->",
		},
		{
			name: "Active reviewers with default template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-2 * time.Hour), IsTeam: true}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template: DefaultTemplate,
			expected: "Awaiting reviews from: @reviewer1, @team1 (team)\n<!-- gong -->",
		},
		{
			name: "Disabled reviewers only",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-45 * time.Minute)}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			template: "No active reviewers to ping.\n<!-- gong -->",
			expected: "No active reviewers to ping.\n<!-- gong -->",
		},
		{
			name: "Mixed reviewers with custom template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-3 * time.Hour), IsTeam: true}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			template: `📌 Please review: {{ range $i, $r := .ActiveReviewers }}{{ if $i }}, {{ end }}@{{ $r }}{{ end }}
{{ if .DisabledReviewers }}
⏳ Not pinging: {{ range $i, $r := .DisabledReviewers }}{{ if $i }}, {{ end }}{{ $r }}{{ end }}
{{ end }}
<!-- gong -->`,
			expected: `📌 Please review: @reviewer1

⏳ Not pinging: reviewer2 (1h ago, delay: 3600s), status: waiting, team1 (team) (3h ago, delay: 3600s), status: disabled

<!-- gong -->`,
		},
		{
			name: "Custom simple template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template: "Active: {{len .ActiveReviewers}}, Inactive: {{len .DisabledReviewers}}\n<!-- gong -->",
			expected: "Active: 1, Inactive: 0\n<!-- gong -->",
		},
		{
			name: "Access to full ping request data",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template: "{{range .PingRequests}}Reviewer: @{{.Req.From}}, Delay: {{.Delay}}s{{end}}\n<!-- gong -->",
			expected: "Reviewer: @reviewer1, Delay: 3600s\n<!-- gong -->",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := formatWithTemplate(tc.requests, tc.template)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if result != tc.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tc.expected, result)
			}
		})
	}
}

// TestTemplateEdgeCases tests how the template handling handles special template cases
func TestTemplateEdgeCases(t *testing.T) {
	testCases := []struct {
		name       string
		template   string
		shouldWork bool // Whether the template should parse and execute without errors
	}{
		{
			name:       "Empty template",
			template:   "",
			shouldWork: true,
		},
		{
			name:       "Just static text",
			template:   "This is a static template with no variables\n<!-- gong -->",
			shouldWork: true,
		},
		{
			name:       "Invalid template syntax",
			template:   "{{.PingRequests}\n<!-- gong -->", // Missing closing brace
			shouldWork: false,
		},
		{
			name:       "Complex template with functions",
			template:   "{{range .PingRequests}}{{if .ShouldPing}}Ping @{{.Req.From}}{{end}}{{end}}\n<!-- gong -->",
			shouldWork: true,
		},
	}

	// Create a simple ping request for testing
	pingRequests := []ping.PingRequest{
		{
			Req:        githubclient.ReviewRequest{From: "tester", On: time.Now()},
			Delay:      3600,
			Enabled:    true,
			ShouldPing: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := formatWithTemplate(pingRequests, tc.template)
			if tc.shouldWork && err != nil {
				t.Errorf("Expected template to work, but got error: %v", err)
			} else if !tc.shouldWork && err == nil {
				t.Errorf("Expected template to fail, but it worked")
			}
		})
	}
}

// TestPrepareTemplateData tests the data preparation for templates
func TestPrepareTemplateData(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name                  string
		requests              []ping.PingRequest
		expectedActiveCount   int
		expectedDisabledCount int
	}{
		{
			name:                  "Empty requests",
			requests:              []ping.PingRequest{},
			expectedActiveCount:   0,
			expectedDisabledCount: 0,
		},
		{
			name: "Only active reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			expectedActiveCount:   2,
			expectedDisabledCount: 0,
		},
		{
			name: "Only disabled reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now}, Enabled: false, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now}, Enabled: true, Delay: 3600, ShouldPing: false},
			},
			expectedActiveCount:   0,
			expectedDisabledCount: 2,
		},
		{
			name: "Mixed reviewers",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now}, Enabled: false, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "reviewer3", On: now}, Enabled: true, Delay: 3600, ShouldPing: false},
			},
			expectedActiveCount:   1,
			expectedDisabledCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := format.PrepareTemplateData(tc.requests, "", "", "", "", true)

			if len(data.ActiveReviewers) != tc.expectedActiveCount {
				t.Errorf("Expected %d active reviewers, got %d", tc.expectedActiveCount, len(data.ActiveReviewers))
			}

			if len(data.DisabledReviewers) != tc.expectedDisabledCount {
				t.Errorf("Expected %d disabled reviewers, got %d", tc.expectedDisabledCount, len(data.DisabledReviewers))
			}

			// Make sure the ping requests are stored in the template data
			if len(data.PingRequests) != len(tc.requests) {
				t.Errorf("Expected %d ping requests, got %d", len(tc.requests), len(data.PingRequests))
			}
		})
	}
}
