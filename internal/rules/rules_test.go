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
		{From: "reviewer1", On: timeNow().Add(-2 * time.Hour)},
		{From: "reviewer2", On: timeNow().Add(-1 * time.Hour)},
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
		{From: "reviewer1", On: timeNow().Add(-2 * time.Hour)},
		{From: "reviewer2", On: timeNow().Add(-1 * time.Hour)},
		{From: "reviewer3", On: timeNow().Add(-3 * time.Hour)},
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
		{From: "reviewer1", On: timeNow().Add(-2 * time.Hour)},
		{From: "reviewer2", On: timeNow().Add(-4 * time.Hour)},
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
						"matchName": "test-user",
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
						"matchName": "test-user",
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
