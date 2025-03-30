package slack

import (
	"context"
	"fmt"
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
		r.Body.Read(buf)
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

func TestSendSlackMessage(t *testing.T) {
	testCases := []struct {
		name           string
		channel        string
		repoOwner      string
		repoName       string
		prNumber       string
		reviewers      []string
		isDryRun       bool
		expectWebhook  bool
		expectContains []string
	}{
		{
			name:          "Dry run mode",
			channel:       "general",
			repoOwner:     "owner",
			repoName:      "repo",
			prNumber:      "123",
			reviewers:     []string{"reviewer1", "reviewer2"},
			isDryRun:      true,
			expectWebhook: false,
		},
		{
			name:          "Single reviewer",
			channel:       "code-reviews",
			repoOwner:     "owner",
			repoName:      "repo",
			prNumber:      "123",
			reviewers:     []string{"reviewer1"},
			isDryRun:      false,
			expectWebhook: true,
			expectContains: []string{
				"code-reviews",
				"reviewer1",
				"PR #123 is waiting for your review",
				"https://github.com/owner/repo/pull/123",
			},
		},
		{
			name:          "Multiple reviewers",
			channel:       "general",
			repoOwner:     "owner",
			repoName:      "repo",
			prNumber:      "456",
			reviewers:     []string{"reviewer1", "reviewer2", "team-name"},
			isDryRun:      false,
			expectWebhook: true,
			expectContains: []string{
				"general",
				"reviewer1, reviewer2, team-name",
				"PR #456 is waiting for your review",
				"https://github.com/owner/repo/pull/456",
			},
		},
		{
			name:          "No reviewers",
			channel:       "general",
			repoOwner:     "owner",
			repoName:      "repo",
			prNumber:      "789",
			reviewers:     []string{},
			isDryRun:      false,
			expectWebhook: true,
			expectContains: []string{
				"PR #789 is waiting for your review",
				"https://github.com/owner/repo/pull/789",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup test server to capture Slack webhook requests
			var capturedRequest []byte
			var serverCalled bool

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverCalled = true
				// Read request body
				buf := make([]byte, r.ContentLength)
				r.Body.Read(buf)
				capturedRequest = buf
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Call the function
			SendSlackMessage(tc.channel, server.URL, tc.repoOwner, tc.repoName, tc.prNumber, tc.reviewers, tc.isDryRun)

			// Verify webhook call
			assert.Equal(t, tc.expectWebhook, serverCalled, "Expected webhook called: %v, got: %v", tc.expectWebhook, serverCalled)

			// If we expect webhook and there are expected contents, verify them
			if tc.expectWebhook && tc.expectContains != nil {
				requestStr := string(capturedRequest)
				for _, expected := range tc.expectContains {
					assert.Contains(t, requestStr, expected, "Expected webhook request to contain '%s'", expected)
				}
			}
		})
	}
}

func TestSendSlackMessageError(t *testing.T) {
	// Setup test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	}))
	defer server.Close()

	// This should not panic even with the error response
	SendSlackMessage(
		"general",
		server.URL,
		"owner",
		"repo",
		"123",
		[]string{"reviewer1"},
		false,
	)

	// No assertion needed - we're testing that it handles errors gracefully
}
