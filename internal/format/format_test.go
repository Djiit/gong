package format

import (
	"testing"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "just now",
			duration: 30 * time.Second,
			want:     "just now",
		},
		{
			name:     "minutes only",
			duration: 45 * time.Minute,
			want:     "45m",
		},
		{
			name:     "hours only",
			duration: 5 * time.Hour,
			want:     "5h",
		},
		{
			name:     "days only",
			duration: 48 * time.Hour,
			want:     "2d",
		},
		{
			name:     "days and hours",
			duration: 50 * time.Hour,
			want:     "2d 2h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPrepareTemplateData(t *testing.T) {
	now := time.Now()
	pingRequests := []ping.PingRequest{
		{
			Req: githubclient.ReviewRequest{
				From:   "user1",
				On:     now.Add(-24 * time.Hour),
				IsTeam: false,
			},
			ShouldPing: true,
			Enabled:    true,
			Delay:      300,
		},
		{
			Req: githubclient.ReviewRequest{
				From:   "team1",
				On:     now.Add(-48 * time.Hour),
				IsTeam: true,
			},
			ShouldPing: false,
			Enabled:    false,
			Delay:      600,
		},
	}

	t.Run("with full info", func(t *testing.T) {
		data := PrepareTemplateData(pingRequests, "owner", "repo", "123", "https://github.com/owner/repo/pull/123", true)

		assert.Len(t, data.ActiveReviewers, 1)
		assert.Contains(t, data.ActiveReviewers[0], "user1")
		assert.Contains(t, data.ActiveReviewers[0], "1d ago")

		assert.Len(t, data.DisabledReviewers, 1)
		assert.Contains(t, data.DisabledReviewers[0], "team1 (team)")
		assert.Contains(t, data.DisabledReviewers[0], "2d ago")
		assert.Contains(t, data.DisabledReviewers[0], "status: disabled")

		assert.Equal(t, "owner", data.RepoOwner)
		assert.Equal(t, "repo", data.RepoName)
		assert.Equal(t, "123", data.PRNumber)
		assert.Equal(t, "https://github.com/owner/repo/pull/123", data.PRURL)
	})

	t.Run("without full info", func(t *testing.T) {
		data := PrepareTemplateData(pingRequests, "owner", "repo", "123", "https://github.com/owner/repo/pull/123", false)

		assert.Len(t, data.ActiveReviewers, 1)
		assert.Equal(t, "user1", data.ActiveReviewers[0])

		assert.Len(t, data.DisabledReviewers, 1)
		assert.Contains(t, data.DisabledReviewers[0], "team1 (team)")
	})
}
