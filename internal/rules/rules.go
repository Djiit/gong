package rules

import (
	"context"
	"path/filepath"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
)

// Variable to allow time.Now to be mocked in tests
var timeNow = time.Now

// Rule represents a rule for matching reviewers with custom delays
type Rule struct {
	MatchName string
	Delay     int
	Enabled   bool // Whether this rule is enabled
}

// Each rule can override the global delay for specific reviewers matching the glob pattern
// It also updates the Delay, Enabled, and ShouldPing fields for each request.
func ApplyRules(ctx context.Context, requests []githubclient.ReviewRequest, rules []Rule) []githubclient.ReviewRequest {
	var resultRequests []githubclient.ReviewRequest
	now := timeNow()

	for _, req := range requests {
		// Create a copy of the request
		updatedReq := req
		updatedReq.Delay = ctx.Value("delay").(int)
		updatedReq.Enabled = ctx.Value("enabled").(bool)

		// Check if any rule matches this reviewer
		for _, rule := range rules {
			if matched, _ := filepath.Match(rule.MatchName, req.From); matched {
				updatedReq.Delay = rule.Delay
				updatedReq.Enabled = rule.Enabled
				break
			}
		}

		// Determine if we should ping based on delay and enabled status
		updatedReq.ShouldPing = updatedReq.Enabled && (updatedReq.Delay <= 0 || now.Sub(req.On).Seconds() >= float64(updatedReq.Delay))

		// Only include in filtered results if should ping
		if updatedReq.ShouldPing {
			resultRequests = append(resultRequests, updatedReq)
		} else {
			// We still want to return the request with the updated status information
			// but it won't be used for actual pinging
			resultRequests = append(resultRequests, updatedReq)
		}
	}

	return resultRequests
}
