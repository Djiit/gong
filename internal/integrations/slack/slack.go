package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/Djiit/gong/internal/ping"
)

func Run(ctx context.Context) {
	pingRequests := ctx.Value("pingRequests").([]ping.PingRequest)
	repoOwner := ctx.Value("repoOwner").(string)
	repoName := ctx.Value("repoName").(string)
	prNumber := ctx.Value("pr").(string)
	isDryRun := ctx.Value("dry-run").(bool)

	if len(pingRequests) == 0 {
		return
	}

	// Extract reviewers to mention
	var reviewers []string
	for _, req := range pingRequests {
		reviewers = append(reviewers, req.Req.From)
	}

	// Find the integration parameters for Slack
	// Use parameters from the first request since all requests in the same group
	// should have the same parameters for a given integration type
	var channel string
	for _, intg := range pingRequests[0].Integrations {
		if intg.Type == "slack" {
			// Look for channel parameter
			if ch, ok := intg.Parameters["channel"]; ok && ch != "" {
				channel = ch
				break
			}
		}
	}

	// If no channel found, use default
	if channel == "" {
		channel = "general" // Default channel if not specified
	}

	if isDryRun {
		fmt.Printf("[DRY RUN] Would send Slack notification to channel %s for PR #%s in %s/%s mentioning: @%s\n",
			channel, prNumber, repoOwner, repoName, strings.Join(reviewers, ", @"))
		return
	}

	// Here would go the actual Slack API integration
	// For now, just print what we would do
	fmt.Printf("Sending Slack notification to channel %s for PR #%s in %s/%s mentioning: @%s\n",
		channel, prNumber, repoOwner, repoName, strings.Join(reviewers, ", @"))
}
