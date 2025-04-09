package rules

import (
	"context"
	"testing"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestApplyRules(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()

	// Create context with global integrations
	globalIntegrations := []ping.Integration{
		{
			Type:       "stdout",
			Parameters: map[string]string{},
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "delay", 0)
	ctx = context.WithValue(ctx, "enabled", true)
	ctx = context.WithValue(ctx, "integrations", globalIntegrations)

	requests := []githubclient.ReviewRequest{
		{From: "reviewer1", On: timeNow().Add(-2 * time.Hour), PRAuthor: "author1"},
		{From: "reviewer2", On: timeNow().Add(-1 * time.Hour), PRAuthor: "author2"},
	}

	rules := []Rule{
		{MatchName: "reviewer1", Delay: 3600, Enabled: true},
		{MatchName: "reviewer2", Delay: 86400, Enabled: true},
	}

	result := ApplyRules(ctx, requests, rules)

	assert.Equal(t, 2, len(result))
	assert.True(t, result[0].ShouldPing)
	assert.False(t, result[1].ShouldPing)
}

func TestApplyRulesWithMultipleIntegrations(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()

	// Create global integrations
	globalIntegrations := []ping.Integration{
		{
			Type:       "stdout",
			Parameters: map[string]string{},
		},
		{
			Type: "slack",
			Parameters: map[string]string{
				"channel": "general",
			},
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "delay", 0)
	ctx = context.WithValue(ctx, "enabled", true)
	ctx = context.WithValue(ctx, "integrations", globalIntegrations)

	requests := []githubclient.ReviewRequest{
		{From: "reviewer1", On: timeNow().Add(-2 * time.Hour), PRAuthor: "author1"},
		{From: "reviewer2", On: timeNow().Add(-1 * time.Hour), PRAuthor: "author2"},
		{From: "reviewer3", On: timeNow().Add(-3 * time.Hour), PRAuthor: "author3"},
	}

	// Rule with custom integrations
	ruleIntegrations := []ping.Integration{
		{
			Type:       "comment",
			Parameters: map[string]string{},
		},
		{
			Type: "slack",
			Parameters: map[string]string{
				"channel": "urgent",
			},
		},
	}

	rules := []Rule{
		{MatchName: "reviewer1", Delay: 3600, Enabled: true, Integrations: ruleIntegrations},
		{MatchName: "reviewer2", Delay: 86400, Enabled: true}, // No integrations specified, should use global
		{MatchName: "reviewer3", Delay: 0, Enabled: true},     // No integrations specified, should use global
	}

	result := ApplyRules(ctx, requests, rules)

	assert.Equal(t, 3, len(result))

	// Check reviewer1 uses rule integrations
	assert.Equal(t, 2, len(result[0].Integrations))
	assert.Equal(t, "comment", result[0].Integrations[0].Type)
	assert.Equal(t, "slack", result[0].Integrations[1].Type)
	assert.Equal(t, "urgent", result[0].Integrations[1].Parameters["channel"])

	// Check reviewer2 uses global integrations
	assert.Equal(t, 2, len(result[1].Integrations))
	assert.Equal(t, "stdout", result[1].Integrations[0].Type)
	assert.Equal(t, "slack", result[1].Integrations[1].Type)
	assert.Equal(t, "general", result[1].Integrations[1].Parameters["channel"])

	// Check shouldPing status
	assert.True(t, result[0].ShouldPing)
	assert.False(t, result[1].ShouldPing)
	assert.True(t, result[2].ShouldPing)
}

func TestEmptyIntegrations(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()

	// Create global integrations
	globalIntegrations := []ping.Integration{
		{
			Type:       "stdout",
			Parameters: map[string]string{},
		},
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "delay", 0)
	ctx = context.WithValue(ctx, "enabled", true)
	ctx = context.WithValue(ctx, "integrations", globalIntegrations)

	requests := []githubclient.ReviewRequest{
		{From: "reviewer1", On: timeNow().Add(-2 * time.Hour), PRAuthor: "author1"},
		{From: "reviewer2", On: timeNow().Add(-4 * time.Hour), PRAuthor: "author2"},
	}

	// Rule with empty integrations list (should use global)
	rules := []Rule{
		{MatchName: "reviewer1", Delay: 3600, Enabled: true, Integrations: []ping.Integration{}},
		{MatchName: "reviewer2", Delay: 0, Enabled: true},
	}

	result := ApplyRules(ctx, requests, rules)

	assert.Equal(t, 2, len(result))

	// Check both reviewers use global integrations
	assert.Equal(t, 1, len(result[0].Integrations))
	assert.Equal(t, "stdout", result[0].Integrations[0].Type)

	assert.Equal(t, 1, len(result[1].Integrations))
	assert.Equal(t, "stdout", result[1].Integrations[0].Type)

	// Check shouldPing status
	assert.True(t, result[0].ShouldPing)
	assert.True(t, result[1].ShouldPing)
}

func TestParseGlobalIntegrations(t *testing.T) {
	tests := []struct {
		name                 string
		config               map[string]interface{}
		expectedIntegrations []ping.Integration
	}{
		{
			name: "Empty Integrations",
			config: map[string]interface{}{
				"integrations": []interface{}{},
			},
			expectedIntegrations: []ping.Integration{},
		},
		{
			name: "Single Integration",
			config: map[string]interface{}{
				"integrations": []interface{}{
					map[string]interface{}{
						"type": "stdout",
					},
				},
			},
			expectedIntegrations: []ping.Integration{
				{
					Type:       "stdout",
					Parameters: map[string]string{},
				},
			},
		},
		{
			name: "Integration With Parameters",
			config: map[string]interface{}{
				"integrations": []interface{}{
					map[string]interface{}{
						"type": "slack",
						"params": map[string]interface{}{
							"channel": "#general",
						},
					},
				},
			},
			expectedIntegrations: []ping.Integration{
				{
					Type: "slack",
					Parameters: map[string]string{
						"channel": "#general",
					},
				},
			},
		},
		{
			name: "Multiple Integrations",
			config: map[string]interface{}{
				"integrations": []interface{}{
					map[string]interface{}{
						"type": "stdout",
					},
					map[string]interface{}{
						"type": "slack",
						"params": map[string]interface{}{
							"channel": "#general",
						},
					},
					map[string]interface{}{
						"type": "comment",
					},
				},
			},
			expectedIntegrations: []ping.Integration{
				{
					Type:       "stdout",
					Parameters: map[string]string{},
				},
				{
					Type: "slack",
					Parameters: map[string]string{
						"channel": "#general",
					},
				},
				{
					Type:       "comment",
					Parameters: map[string]string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			for k, v := range tt.config {
				viper.Set(k, v)
			}

			result := ParseGlobalIntegrations()

			assert.Equal(t, len(tt.expectedIntegrations), len(result))

			if len(tt.expectedIntegrations) > 0 {
				for i, expected := range tt.expectedIntegrations {
					assert.Equal(t, expected.Type, result[i].Type)

					// Check parameters
					assert.Equal(t, len(expected.Parameters), len(result[i].Parameters))
					for k, v := range expected.Parameters {
						assert.Equal(t, v, result[i].Parameters[k])
					}
				}
			}
		})
	}
}

func TestParseRules(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]interface{}
		expectedRules  []Rule
		expectedLength int
	}{
		{
			name: "Empty Rules",
			config: map[string]interface{}{
				"rules": []interface{}{},
			},
			expectedRules:  []Rule{},
			expectedLength: 0,
		},
		{
			name: "Rule With Single Integration",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchname": "test-user",
						"delay":     24,
						"enabled":   true,
						"integrations": []interface{}{
							map[string]interface{}{
								"type": "stdout",
							},
						},
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName: "test-user",
					Delay:     24,
					Enabled:   true,
					Integrations: []ping.Integration{
						{
							Type:       "stdout",
							Parameters: map[string]string{},
						},
					},
				},
			},
			expectedLength: 1,
		},
		{
			name: "Rule With Multiple Integrations",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchname": "test-user",
						"delay":     24,
						"enabled":   true,
						"integrations": []interface{}{
							map[string]interface{}{
								"type": "slack",
								"params": map[string]interface{}{
									"channel": "#urgent",
								},
							},
							map[string]interface{}{
								"type": "comment",
							},
						},
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName: "test-user",
					Delay:     24,
					Enabled:   true,
					Integrations: []ping.Integration{
						{
							Type: "slack",
							Parameters: map[string]string{
								"channel": "#urgent",
							},
						},
						{
							Type:       "comment",
							Parameters: map[string]string{},
						},
					},
				},
			},
			expectedLength: 1,
		},
		{
			name: "Rule Without MatchName",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"delay":   24,
						"enabled": true,
						"integrations": []interface{}{
							map[string]interface{}{
								"type": "stdout",
							},
						},
					},
				},
			},
			expectedRules:  []Rule{},
			expectedLength: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			for k, v := range tt.config {
				viper.Set(k, v)
			}

			result := ParseRules()
			assert.Equal(t, tt.expectedLength, len(result))

			if tt.expectedLength > 0 {
				for i, rule := range tt.expectedRules {
					assert.Equal(t, rule.MatchName, result[i].MatchName)
					assert.Equal(t, rule.Delay, result[i].Delay)
					assert.Equal(t, rule.Enabled, result[i].Enabled)

					// Check integrations
					assert.Equal(t, len(rule.Integrations), len(result[i].Integrations))

					for j, integration := range rule.Integrations {
						assert.Equal(t, integration.Type, result[i].Integrations[j].Type)

						// Check parameters
						assert.Equal(t, len(integration.Parameters), len(result[i].Integrations[j].Parameters))
						for k, v := range integration.Parameters {
							assert.Equal(t, v, result[i].Integrations[j].Parameters[k])
						}
					}
				}
			}
		})
	}
}

func TestMatchTitleRule(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()

	// Create a context with global settings
	ctx := context.Background()
	ctx = context.WithValue(ctx, "delay", 0)
	ctx = context.WithValue(ctx, "enabled", true)
	ctx = context.WithValue(ctx, "integrations", []ping.Integration{
		{Type: "stdout", Parameters: map[string]string{}},
	})

	// Create test review requests with different PR titles
	requests := []githubclient.ReviewRequest{
		{From: "reviewer1", PRTitle: "feat: add new feature", PRAuthor: "author1", On: timeNow().Add(-2 * time.Hour)},
		{From: "reviewer1", PRTitle: "fix: bug fix", PRAuthor: "author1", On: timeNow().Add(-2 * time.Hour)},
		{From: "reviewer2", PRTitle: "docs: update documentation", PRAuthor: "author2", On: timeNow().Add(-1 * time.Hour)},
		{From: "reviewer3", PRTitle: "chore: update dependencies", PRAuthor: "author3", On: timeNow().Add(-3 * time.Hour)},
	}

	// Test rules for matchTitle
	tests := []struct {
		name         string
		rules        []Rule
		expectations map[string]bool // Maps PR title to expected ShouldPing value
	}{
		{
			name: "Match by title pattern only",
			rules: []Rule{
				{MatchTitle: "feat: *", Delay: 3600, Enabled: true},
				{MatchTitle: "fix: *", Delay: 86400, Enabled: true},
			},
			expectations: map[string]bool{
				"feat: add new feature":      true,  // Matched first rule, delay < time passed
				"fix: bug fix":               false, // Matched second rule, delay > time passed
				"docs: update documentation": true,  // Not matched by any rule, uses global delay (0)
				"chore: update dependencies": true,  // Not matched by any rule, uses global delay (0)
			},
		},
		{
			name: "Match by both name and title",
			rules: []Rule{
				{MatchName: "reviewer1", MatchTitle: "feat: *", Delay: 3600, Enabled: true},
				{MatchName: "reviewer2", MatchTitle: "docs: *", Delay: 0, Enabled: false},
			},
			expectations: map[string]bool{
				"feat: add new feature":      true,  // Matches first rule, delay < time passed
				"fix: bug fix":               true,  // reviewer1 but title doesn't match, uses global (enabled,delay=0)
				"docs: update documentation": false, // Matches second rule (disabled)
				"chore: update dependencies": true,  // Not matched by any rule, uses global settings
			},
		},
		{
			name: "Multiple title patterns",
			rules: []Rule{
				{MatchTitle: "feat: *", Delay: 3600, Enabled: true},
				{MatchTitle: "fix: *", Delay: 86400, Enabled: true},
				{MatchTitle: "docs: *", Delay: 0, Enabled: false},
				{MatchTitle: "chore: *", Delay: 0, Enabled: true},
			},
			expectations: map[string]bool{
				"feat: add new feature":      true,  // Matched by first rule
				"fix: bug fix":               false, // Matched by second rule
				"docs: update documentation": false, // Matched by third rule (disabled)
				"chore: update dependencies": true,  // Matched by fourth rule
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyRules(ctx, requests, tt.rules)

			// Check if each result has the expected ShouldPing value based on its PRTitle
			for _, r := range result {
				expectedShouldPing, exists := tt.expectations[r.Req.PRTitle]
				if !exists {
					t.Errorf("No expectation found for PR title %q", r.Req.PRTitle)
					continue
				}
				assert.Equal(t, expectedShouldPing, r.ShouldPing, "For PR title %q", r.Req.PRTitle)
			}
		})
	}
}

func TestParseRulesWithMatchTitle(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedRules []Rule
	}{
		{
			name: "Rules with matchTitle",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchtitle": "feat: *",
						"delay":      24,
						"enabled":    true,
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchTitle: "feat: *",
					Delay:      24,
					Enabled:    true,
				},
			},
		},
		{
			name: "Rules with both matchName and matchTitle",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchname":  "test-user",
						"matchtitle": "fix: *",
						"delay":      48,
						"enabled":    true,
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName:  "test-user",
					MatchTitle: "fix: *",
					Delay:      48,
					Enabled:    true,
				},
			},
		},
		{
			name: "Rules with only matchTitle but no matchName",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchtitle": "docs: *",
						"delay":      12,
						"enabled":    false,
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchTitle: "docs: *",
					Delay:      12,
					Enabled:    false,
				},
			},
		},
		{
			name: "Mixed rules with different match conditions",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchname": "user1",
						"delay":     24,
						"enabled":   true,
					},
					map[string]interface{}{
						"matchtitle": "feat: *",
						"delay":      36,
						"enabled":    true,
					},
					map[string]interface{}{
						"matchname":  "user2",
						"matchtitle": "fix: *",
						"delay":      48,
						"enabled":    false,
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName: "user1",
					Delay:     24,
					Enabled:   true,
				},
				{
					MatchTitle: "feat: *",
					Delay:      36,
					Enabled:    true,
				},
				{
					MatchName:  "user2",
					MatchTitle: "fix: *",
					Delay:      48,
					Enabled:    false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			for k, v := range tt.config {
				viper.Set(k, v)
			}

			result := ParseRules()
			assert.Equal(t, len(tt.expectedRules), len(result))

			for i, rule := range tt.expectedRules {
				assert.Equal(t, rule.MatchName, result[i].MatchName)
				assert.Equal(t, rule.MatchTitle, result[i].MatchTitle)
				assert.Equal(t, rule.Delay, result[i].Delay)
				assert.Equal(t, rule.Enabled, result[i].Enabled)
			}
		})
	}
}

func TestMatchAuthorRule(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()

	// Create a context with global settings
	ctx := context.Background()
	ctx = context.WithValue(ctx, "delay", 0)
	ctx = context.WithValue(ctx, "enabled", true)
	ctx = context.WithValue(ctx, "integrations", []ping.Integration{
		{Type: "stdout", Parameters: map[string]string{}},
	})

	// Create test review requests with different PR authors
	requests := []githubclient.ReviewRequest{
		{From: "reviewer1", PRAuthor: "author1", On: timeNow().Add(-2 * time.Hour)},
		{From: "reviewer1", PRAuthor: "author2", On: timeNow().Add(-2 * time.Hour)},
		{From: "reviewer2", PRAuthor: "author1", On: timeNow().Add(-1 * time.Hour)},
		{From: "reviewer3", PRAuthor: "author3", On: timeNow().Add(-3 * time.Hour)},
	}

	// Test rules for matchAuthor
	tests := []struct {
		name         string
		rules        []Rule
		expectations map[string]bool // Maps PR author to expected ShouldPing value
	}{
		{
			name: "Match by author pattern only",
			rules: []Rule{
				{MatchAuthor: "author1", Delay: 3600, Enabled: true},
				{MatchAuthor: "author2", Delay: 86400, Enabled: true},
			},
			expectations: map[string]bool{
				"author1": true,  // Matched first rule, delay < time passed
				"author2": false, // Matched second rule, delay > time passed
				"author3": true,  // Not matched by any rule, uses global delay (0)
			},
		},
		{
			name: "Multiple match conditions",
			rules: []Rule{
				{MatchName: "reviewer1", MatchAuthor: "author1", Delay: 3600, Enabled: true},
				{MatchAuthor: "author2", Delay: 86400, Enabled: true},
				{MatchTitle: "some-title", MatchAuthor: "author3", Delay: 0, Enabled: false},
			},
			expectations: map[string]bool{
				"author1": true,  // Matched by first rule for reviewer1
				"author2": false, // Matched by second rule
				"author3": true,  // Title doesn't match the third rule, uses global settings
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyRules(ctx, requests, tt.rules)

			// Check if each result has the expected ShouldPing value based on its PRAuthor
			for _, r := range result {
				expectedShouldPing, exists := tt.expectations[r.Req.PRAuthor]
				if !exists {
					t.Errorf("No expectation found for PR author %q", r.Req.PRAuthor)
					continue
				}

				// For the case where reviewer2 and author1 should be disabled
				if r.Req.From == "reviewer2" && r.Req.PRAuthor == "author1" && tt.name == "Match by both name and author" {
					assert.False(t, r.ShouldPing, "For reviewer %q and author %q", r.Req.From, r.Req.PRAuthor)
				} else {
					assert.Equal(t, expectedShouldPing, r.ShouldPing, "For reviewer %q and author %q", r.Req.From, r.Req.PRAuthor)
				}
			}
		})
	}
}

func TestParseRulesWithMatchAuthor(t *testing.T) {
	tests := []struct {
		name          string
		config        map[string]interface{}
		expectedRules []Rule
	}{
		{
			name: "Rules with matchAuthor",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchauthor": "author1", // Updated to camelCase to match config
						"delay":       24,
						"enabled":     true,
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchAuthor: "author1",
					Delay:       24,
					Enabled:     true,
				},
			},
		},
		{
			name: "Rules with both matchName and matchAuthor",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchname":   "test-user", // Updated to camelCase
						"matchauthor": "author1",   // Updated to camelCase
						"delay":       48,
						"enabled":     true,
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName:   "test-user",
					MatchAuthor: "author1",
					Delay:       48,
					Enabled:     true,
				},
			},
		},
		{
			name: "Rules with multiple match conditions",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchname":   "reviewer1", // Updated to camelCase
						"matchtitle":  "feat: *",   // Updated to camelCase
						"matchauthor": "author1",   // Updated to camelCase
						"delay":       36,
						"enabled":     true,
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName:   "reviewer1",
					MatchTitle:  "feat: *",
					MatchAuthor: "author1",
					Delay:       36,
					Enabled:     true,
				},
			},
		},
		{
			name: "Mixed rules with different match conditions",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchname": "reviewer1", // Updated to camelCase
						"delay":     24,
						"enabled":   true,
					},
					map[string]interface{}{
						"matchauthor": "author1", // Updated to camelCase
						"delay":       36,
						"enabled":     true,
					},
					map[string]interface{}{
						"matchtitle":  "fix: *",  // Updated to camelCase
						"matchauthor": "author2", // Updated to camelCase
						"delay":       48,
						"enabled":     false,
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName: "reviewer1",
					Delay:     24,
					Enabled:   true,
				},
				{
					MatchAuthor: "author1",
					Delay:       36,
					Enabled:     true,
				},
				{
					MatchTitle:  "fix: *",
					MatchAuthor: "author2",
					Delay:       48,
					Enabled:     false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			for k, v := range tt.config {
				viper.Set(k, v)
			}

			result := ParseRules()
			assert.Equal(t, len(tt.expectedRules), len(result))

			for i, rule := range tt.expectedRules {
				assert.Equal(t, rule.MatchName, result[i].MatchName)
				assert.Equal(t, rule.MatchTitle, result[i].MatchTitle)
				assert.Equal(t, rule.MatchAuthor, result[i].MatchAuthor)
				assert.Equal(t, rule.Delay, result[i].Delay)
				assert.Equal(t, rule.Enabled, result[i].Enabled)
			}
		})
	}
}
