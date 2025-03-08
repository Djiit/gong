package stdout

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Djiit/gong/internal/format"
	"github.com/Djiit/gong/internal/ping"
)

func Run(ctx context.Context) {
	pingRequests := ctx.Value("pingRequests").([]ping.PingRequest)
	isDryRun := ctx.Value("dry-run").(bool)

	if isDryRun {
		fmt.Println("[DRY RUN] Would output reviewer information to stdout")
		return
	}

	output := formatPingRequests(pingRequests)
	fmt.Println(output)
}

func formatPingRequests(pingRequests []ping.PingRequest) string {
	if len(pingRequests) == 0 {
		return "No pending review requests."
	}

	var activeReviewers []string
	var disabledReviewers []string

	for _, req := range pingRequests {
		timeSinceRequest := time.Since(req.Req.On).Round(time.Hour)
		formattedDuration := format.FormatDuration(timeSinceRequest)

		reviewer := req.Req.From
		if req.Req.IsTeam {
			reviewer += " (team)"
		}

		reviewerInfo := fmt.Sprintf("%s (%s ago, delay: %ds)",
			reviewer, formattedDuration, req.Delay)

		if req.ShouldPing {
			activeReviewers = append(activeReviewers, reviewerInfo)
		} else {
			status := "waiting"
			if !req.Enabled {
				status = "disabled"
			}
			disabledReviewers = append(disabledReviewers, fmt.Sprintf("%s, status: %s", reviewerInfo, status))
		}
	}

	var result strings.Builder

	if len(activeReviewers) > 0 {
		result.WriteString(fmt.Sprintf("Pinging: %s\n", strings.Join(activeReviewers, ", ")))
	}

	if len(disabledReviewers) > 0 {
		result.WriteString(fmt.Sprintf("Not pinging: %s", strings.Join(disabledReviewers, ", ")))
	}

	return result.String()
}
