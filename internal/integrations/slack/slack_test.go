package slack

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	// Setup test server to capture Slack webhook requests
	var capturedRequest []byte
	var serverCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		// Read request body
		buf := make([]byte, r.ContentLength)
		_, err := r.Body.Read(buf)
		if err != nil && err != io.EOF {
			t.Fatal(err)
		}
		capturedRequest = buf
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Reset for each test
	viper.Reset()
	viper.Set("slack-webhook", server.URL)

	now := time.Now()

	testCases := []struct {
		name           string
		pingRequests   []ping.PingRequest
		repoOwner      string
		repoName       string
		prNumber       string
		isDryRun       bool
		expectWebhook  bool
		expectContains string
	}{
		{
			name:          "No ping requests",
			pingRequests:  []ping.PingRequest{},
			repoOwner:     "owner",
			repoName:      "repo",
			prNumber:      "123",
			isDryRun:      false,
			expectWebhook: false,
		},
		{
			name: "With ping requests but dry run",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type: "slack",
							Parameters: map[string]string{
								"channel": "code-reviews",
							},
						},
					},
				},
			},
			repoOwner:     "owner",
			repoName:      "repo",
			prNumber:      "123",
			isDryRun:      true,
			expectWebhook: false,
		},
		{
			name: "With ping requests and custom channel",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type: "slack",
							Parameters: map[string]string{
								"channel": "code-reviews",
							},
						},
					},
				},
			},
			repoOwner:      "owner",
			repoName:       "repo",
			prNumber:       "123",
			isDryRun:       false,
			expectWebhook:  true,
			expectContains: "code-reviews",
		},
		{
			name: "With ping requests and default channel",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type:       "slack",
							Parameters: map[string]string{},
						},
					},
				},
			},
			repoOwner:      "owner",
			repoName:       "repo",
			prNumber:       "123",
			isDryRun:       false,
			expectWebhook:  true,
			expectContains: "general",
		},
		{
			name: "With multiple reviewers",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type: "slack",
							Parameters: map[string]string{
								"channel": "code-reviews",
							},
						},
					},
				},
				{
					Req:        githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-3 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type: "slack",
							Parameters: map[string]string{
								"channel": "code-reviews",
							},
						},
					},
				},
			},
			repoOwner:      "owner",
			repoName:       "repo",
			prNumber:       "123",
			isDryRun:       false,
			expectWebhook:  true,
			expectContains: "reviewer1, reviewer2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state for each test
			serverCalled = false
			capturedRequest = nil

			// Create context with values
			ctx := context.Background()
			ctx = context.WithValue(ctx, "pingRequests", tc.pingRequests)
			ctx = context.WithValue(ctx, "repoOwner", tc.repoOwner)
			ctx = context.WithValue(ctx, "repoName", tc.repoName)
			ctx = context.WithValue(ctx, "pr", tc.prNumber)
			ctx = context.WithValue(ctx, "dry-run", tc.isDryRun)

			Run(ctx)

			// Verify webhook call
			assert.Equal(t, tc.expectWebhook, serverCalled, "Expected webhook called: %v, got: %v", tc.expectWebhook, serverCalled)

			// If we expect webhook and there are messages, verify content
			if tc.expectWebhook && tc.expectContains != "" {
				assert.Contains(t, string(capturedRequest), tc.expectContains, "Expected webhook request to contain '%s'", tc.expectContains)
			}
		})
	}
}

func TestFormatWithTemplate(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name       string
		requests   []ping.PingRequest
		template   string
		repoOwner  string
		repoName   string
		prNumber   string
		prURL      string
		expected   string
		shouldWork bool
	}{
		{
			name:       "No reviewers with default template",
			requests:   []ping.PingRequest{},
			template:   DefaultTemplate,
			repoOwner:  "owner",
			repoName:   "repo",
			prNumber:   "123",
			prURL:      "https://github.com/owner/repo/pull/123",
			expected:   "No pending review requests.",
			shouldWork: true,
		},
		{
			name: "Active reviewers with default template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-2 * time.Hour), IsTeam: true}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template:   DefaultTemplate,
			repoOwner:  "owner",
			repoName:   "repo",
			prNumber:   "123",
			prURL:      "https://github.com/owner/repo/pull/123",
			expected:   "PR #123 is waiting for review: <https://github.com/owner/repo/pull/123|owner/repo#123>\nReviewers: reviewer1, team1 (team)",
			shouldWork: true,
		},
		{
			name: "Disabled reviewers only",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-45 * time.Minute)}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			template:   "No active reviewers to ping.",
			repoOwner:  "owner",
			repoName:   "repo",
			prNumber:   "123",
			prURL:      "https://github.com/owner/repo/pull/123",
			expected:   "No active reviewers to ping.",
			shouldWork: true,
		},
		{
			name: "Mixed reviewers with custom template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-3 * time.Hour), IsTeam: true}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			template: `📌 Please review PR #{{.PRNumber}}: {{ range $i, $r := .ActiveReviewers }}{{ if $i }}, {{ end }}{{ $r }}{{ end }}
{{ if .DisabledReviewers }}
⏳ Not pinging: {{ range $i, $r := .DisabledReviewers }}{{ if $i }}, {{ end }}{{ $r }}{{ end }}
{{ end }}`,
			repoOwner:  "owner",
			repoName:   "repo",
			prNumber:   "123",
			prURL:      "https://github.com/owner/repo/pull/123",
			expected:   "📌 Please review PR #123: reviewer1\n\n⏳ Not pinging: reviewer2 (1h ago, delay: 3600s), status: waiting, team1 (team) (3h ago, delay: 3600s), status: disabled\n",
			shouldWork: true,
		},
		{
			name: "Custom template with URL",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template:   "PR: <{{.PRURL}}|{{.RepoOwner}}/{{.RepoName}}#{{.PRNumber}}>",
			repoOwner:  "owner",
			repoName:   "repo",
			prNumber:   "123",
			prURL:      "https://github.com/owner/repo/pull/123",
			expected:   "PR: <https://github.com/owner/repo/pull/123|owner/repo#123>",
			shouldWork: true,
		},
		{
			name: "Invalid template syntax",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template:   "{{.PingRequests}", // Missing closing brace
			repoOwner:  "owner",
			repoName:   "repo",
			prNumber:   "123",
			prURL:      "https://github.com/owner/repo/pull/123",
			shouldWork: false,
		},
		{
			name: "Access to full ping request data",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template:   "{{range .PingRequests}}Reviewer: {{.Req.From}}, Delay: {{.Delay}}s{{end}}",
			repoOwner:  "owner",
			repoName:   "repo",
			prNumber:   "123",
			prURL:      "https://github.com/owner/repo/pull/123",
			expected:   "Reviewer: reviewer1, Delay: 3600s",
			shouldWork: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := formatWithTemplate(tc.requests, tc.template, tc.repoOwner, tc.repoName, tc.prNumber, tc.prURL)

			if tc.shouldWork {
				assert.NoError(t, err, "Expected no error")
				assert.Equal(t, tc.expected, result, "Template result doesn't match expected output")
			} else {
				assert.Error(t, err, "Expected an error with invalid template")
			}
		})
	}
}

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
			data := prepareTemplateData(tc.requests, "owner", "repo", "123", "https://github.com/owner/repo/pull/123")

			assert.Equal(t, tc.expectedActiveCount, len(data.ActiveReviewers), "Expected %d active reviewers, got %d", tc.expectedActiveCount, len(data.ActiveReviewers))
			assert.Equal(t, tc.expectedDisabledCount, len(data.DisabledReviewers), "Expected %d disabled reviewers, got %d", tc.expectedDisabledCount, len(data.DisabledReviewers))
			assert.Equal(t, len(tc.requests), len(data.PingRequests), "Expected %d ping requests, got %d", len(tc.requests), len(data.PingRequests))

			// Check PR metadata
			assert.Equal(t, "owner", data.RepoOwner)
			assert.Equal(t, "repo", data.RepoName)
			assert.Equal(t, "123", data.PRNumber)
			assert.Equal(t, "https://github.com/owner/repo/pull/123", data.PRURL)
		})
	}
}

func TestRunWithTemplate(t *testing.T) {
	// Setup test server to capture Slack webhook requests
	var capturedRequest []byte
	var serverCalled bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverCalled = true
		// Read request body
		buf := make([]byte, r.ContentLength)
		_, err := r.Body.Read(buf)
		if err != nil && err != io.EOF {
			t.Fatal(err)
		}
		capturedRequest = buf
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Reset for each test
	viper.Reset()
	viper.Set("slack-webhook", server.URL)

	now := time.Now()

	testCases := []struct {
		name           string
		pingRequests   []ping.PingRequest
		template       string
		expectContains string
	}{
		{
			name: "With custom template in integration parameters",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type: "slack",
							Parameters: map[string]string{
								"channel":  "code-reviews",
								"template": "Custom template: {{.PRNumber}} - {{range .ActiveReviewers}}@{{.}} {{end}}",
							},
						},
					},
				},
			},
			expectContains: "Custom template: 123 - @reviewer1",
		},
		{
			name: "With template in context",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type: "slack",
							Parameters: map[string]string{
								"channel": "code-reviews",
							},
						},
					},
				},
			},
			template:       "Context template: {{.PRNumber}} by {{.RepoOwner}}",
			expectContains: "Context template: 123 by owner",
		},
		{
			name: "Falling back to default template",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type:       "slack",
							Parameters: map[string]string{},
						},
					},
				},
			},
			expectContains: "PR #123 is waiting for review:",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset state for each test
			serverCalled = false
			capturedRequest = nil

			// Create context with values
			ctx := context.Background()
			ctx = context.WithValue(ctx, "pingRequests", tc.pingRequests)
			ctx = context.WithValue(ctx, "repoOwner", "owner")
			ctx = context.WithValue(ctx, "repoName", "repo")
			ctx = context.WithValue(ctx, "pr", "123")
			ctx = context.WithValue(ctx, "dry-run", false)

			if tc.template != "" {
				ctx = context.WithValue(ctx, "template", tc.template)
			}

			Run(ctx)

			// Verify webhook call
			assert.True(t, serverCalled, "Expected webhook called")

			// Verify content
			assert.Contains(t, string(capturedRequest), tc.expectContains,
				"Expected webhook request to contain '%s'", tc.expectContains)
		})
	}
}
