package format

import (
	"fmt"
	"time"

	"github.com/Djiit/gong/internal/ping"
)

// FormatDuration formats the duration in a human-readable way
func FormatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24

	if days > 0 {
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}

	minutes := int(d.Minutes()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	}

	return "just now"
}

// TemplateData holds the data for template rendering across all integrations
type TemplateData struct {
	PingRequests      []ping.PingRequest
	ActiveReviewers   []string
	DisabledReviewers []string
	// Additional fields used by specific integrations
	PRNumber  string
	RepoOwner string
	RepoName  string
	PRURL     string
}

// PrepareTemplateData prepares template data from ping requests and optional PR metadata
func PrepareTemplateData(pingRequests []ping.PingRequest, repoOwner, repoName, prNumber, prURL string, includeFullInfo bool) TemplateData {
	var activeReviewers []string
	var disabledReviewers []string

	for _, req := range pingRequests {
		timeSinceRequest := time.Since(req.Req.On).Round(time.Hour)
		formattedDuration := FormatDuration(timeSinceRequest)

		reviewer := req.Req.From
		if req.Req.IsTeam {
			reviewer += " (team)"
		}

		reviewerInfo := fmt.Sprintf("%s (%s ago, delay: %ds)",
			reviewer, formattedDuration, req.Delay)

		if req.ShouldPing {
			// For stdout we want the full info, for others just the name
			if includeFullInfo {
				activeReviewers = append(activeReviewers, reviewerInfo)
			} else {
				activeReviewers = append(activeReviewers, reviewer)
			}
		} else {
			status := "waiting"
			if !req.Enabled {
				status = "disabled"
			}
			disabledReviewers = append(disabledReviewers, fmt.Sprintf("%s, status: %s", reviewerInfo, status))
		}
	}

	return TemplateData{
		PingRequests:      pingRequests,
		ActiveReviewers:   activeReviewers,
		DisabledReviewers: disabledReviewers,
		PRNumber:          prNumber,
		RepoOwner:         repoOwner,
		RepoName:          repoName,
		PRURL:             prURL,
	}
}
