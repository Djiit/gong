package rules

import (
	"context"
	"testing"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/stretchr/testify/assert"
)

func TestApplyRules(t *testing.T) {
	timeNow = func() time.Time { return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { timeNow = time.Now }()

	ctx := context.Background()
	ctx = context.WithValue(ctx, "delay", 0)
	ctx = context.WithValue(ctx, "enabled", true)

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
