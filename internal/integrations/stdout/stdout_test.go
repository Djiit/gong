package stdout

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

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
			expected: "No pending review requests.",
		},
		{
			name: "Active reviewers with default template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-2 * time.Hour), IsTeam: true}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template: DefaultTemplate,
			expected: "Pinging: reviewer1 (1h ago, delay: 3600s), team1 (team) (2h ago, delay: 3600s)",
		},
		{
			name: "Disabled reviewers with default template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-45 * time.Minute)}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			template: DefaultTemplate,
			expected: "Not pinging: reviewer1 (1h ago, delay: 3600s), status: waiting, reviewer2 (1h ago, delay: 3600s), status: disabled",
		},
		{
			name: "Mixed reviewers with default template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-2 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
				{Req: githubclient.ReviewRequest{From: "reviewer2", On: now.Add(-30 * time.Minute)}, Enabled: true, Delay: 3600, ShouldPing: false},
				{Req: githubclient.ReviewRequest{From: "team1", On: now.Add(-3 * time.Hour), IsTeam: true}, Enabled: false, Delay: 3600, ShouldPing: false},
			},
			template: DefaultTemplate,
			expected: "Pinging: reviewer1 (2h ago, delay: 3600s)\nNot pinging: reviewer2 (1h ago, delay: 3600s), status: waiting, team1 (team) (3h ago, delay: 3600s), status: disabled",
		},
		{
			name: "Custom simple template",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template: "Active: {{len .ActiveReviewers}}, Inactive: {{len .DisabledReviewers}}",
			expected: "Active: 1, Inactive: 0",
		},
		{
			name: "Access to full ping request data",
			requests: []ping.PingRequest{
				{Req: githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)}, Enabled: true, Delay: 3600, ShouldPing: true},
			},
			template: "{{range .PingRequests}}Reviewer: {{.Req.From}}, Delay: {{.Delay}}s{{end}}",
			expected: "Reviewer: reviewer1, Delay: 3600s",
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
			template:   "This is a static template with no variables",
			shouldWork: true,
		},
		{
			name:       "Invalid template syntax",
			template:   "{{.PingRequests}", // Missing closing brace
			shouldWork: false,
		},
		{
			name:       "Complex template with functions",
			template:   "{{range .PingRequests}}{{if .ShouldPing}}Ping {{.Req.From}}{{end}}{{end}}",
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

// TestRun tests the Run function with different template scenarios
func TestRun(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name          string
		pingRequests  []ping.PingRequest
		ctxTemplate   string
		integTemplate string
		expected      string
		isDryRun      bool
	}{
		{
			name: "Template from integration parameters",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type: "stdout",
							Parameters: map[string]string{
								"template": "Custom template: {{range .ActiveReviewers}}{{.}}{{end}}",
							},
						},
					},
				},
			},
			ctxTemplate: "",
			expected:    "Custom template: reviewer1 (1h ago, delay: 3600s)",
			isDryRun:    false,
		},
		{
			name: "Template from context",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type:       "stdout",
							Parameters: map[string]string{},
						},
					},
				},
			},
			ctxTemplate: "Context template: {{len .ActiveReviewers}} active",
			expected:    "Context template: 1 active",
			isDryRun:    false,
		},
		{
			name: "Default template when none provided",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
					Integrations: []ping.Integration{
						{
							Type:       "stdout",
							Parameters: map[string]string{},
						},
					},
				},
			},
			ctxTemplate: "",
			expected:    "Pinging: reviewer1 (1h ago, delay: 3600s)",
			isDryRun:    false,
		},
		{
			name: "Dry run mode",
			pingRequests: []ping.PingRequest{
				{
					Req:        githubclient.ReviewRequest{From: "reviewer1", On: now.Add(-1 * time.Hour)},
					Enabled:    true,
					Delay:      3600,
					ShouldPing: true,
				},
			},
			ctxTemplate: "",
			expected:    "[DRY RUN] Would output reviewer information to stdout",
			isDryRun:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create context with test values
			ctx := context.Background()
			ctx = context.WithValue(ctx, "pingRequests", tc.pingRequests)
			ctx = context.WithValue(ctx, "dry-run", tc.isDryRun)
			if tc.ctxTemplate != "" {
				ctx = context.WithValue(ctx, "template", tc.ctxTemplate)
			}

			// Run the function
			Run(ctx)

			// Restore stdout and get captured output
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			os.Stdout = oldStdout
			var buf strings.Builder
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatal(err)
			}
			output := strings.TrimSpace(buf.String())

			// Check output
			if !strings.Contains(output, tc.expected) {
				t.Errorf("expected output to contain %q, got %q", tc.expected, output)
			}
		})
	}
}
