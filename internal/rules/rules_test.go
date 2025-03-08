package rules

import (
	"context"
	"testing"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestApplyRules(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "delay", 0)
	ctx = context.WithValue(ctx, "enabled", true)
	ctx = context.WithValue(ctx, "integration", "comment")

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

func TestApplyRulesWithIntegration(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "delay", 0)
	ctx = context.WithValue(ctx, "enabled", true)
	ctx = context.WithValue(ctx, "integration", "comment")

	requests := []githubclient.ReviewRequest{
		{From: "reviewer1", On: timeNow().Add(-2 * time.Hour)},
		{From: "reviewer2", On: timeNow().Add(-1 * time.Hour)},
		{From: "reviewer3", On: timeNow().Add(-3 * time.Hour)},
	}

	rules := []Rule{
		{MatchName: "reviewer1", Delay: 3600, Enabled: true, Integration: "stdout"},
		{MatchName: "reviewer2", Delay: 86400, Enabled: true}, // No integration specified, should use default
		{MatchName: "reviewer3", Delay: 0, Enabled: true, Integration: "comment"},
	}

	result := ApplyRules(ctx, requests, rules)

	assert.Equal(t, 3, len(result))
	// Check integration overrides
	assert.Equal(t, "stdout", result[0].Integration)
	assert.Equal(t, "comment", result[1].Integration) // Default from context
	assert.Equal(t, "comment", result[2].Integration)

	// Check other fields are still properly set
	assert.True(t, result[0].ShouldPing)
	assert.False(t, result[1].ShouldPing)
	assert.True(t, result[2].ShouldPing)
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
			name: "Single Rule",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchName":   "test-user",
						"delay":       24,
						"enabled":     true,
						"integration": "stdout",
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName:   "test-user",
					Delay:       24,
					Enabled:     true,
					Integration: "stdout",
				},
			},
			expectedLength: 1,
		},
		{
			name: "Multiple Rules",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"matchName":   "test-user1",
						"delay":       24,
						"enabled":     true,
						"integration": "stdout",
					},
					map[string]interface{}{
						"matchName":   "test-user2",
						"delay":       48,
						"enabled":     false,
						"integration": "comment",
					},
				},
			},
			expectedRules: []Rule{
				{
					MatchName:   "test-user1",
					Delay:       24,
					Enabled:     true,
					Integration: "stdout",
				},
				{
					MatchName:   "test-user2",
					Delay:       48,
					Enabled:     false,
					Integration: "comment",
				},
			},
			expectedLength: 2,
		},
		{
			name: "Rule Without MatchName",
			config: map[string]interface{}{
				"rules": []interface{}{
					map[string]interface{}{
						"delay":       24,
						"enabled":     true,
						"integration": "stdout",
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
					assert.Equal(t, rule.Integration, result[i].Integration)
				}
			}
		})
	}
}
