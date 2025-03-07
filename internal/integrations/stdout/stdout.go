package stdout

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Djiit/gong/internal/format"
	"github.com/Djiit/gong/internal/githubclient"
)

func Run(ctx context.Context) {
	reviewRequests := ctx.Value("reviewRequests").([]githubclient.ReviewRequest)
	isDryRun := ctx.Value("dry-run").(bool)

	if isDryRun {
		fmt.Println("[DRY RUN] Would output reviewer information to stdout")
		return
	}

	output := formatReviewRequests(reviewRequests)
	fmt.Println(output)
}

func formatReviewRequests(reviewRequests []githubclient.ReviewRequest) string {
	if len(reviewRequests) == 0 {
		return "No pending review requests."
	}

	var activeReviewers []string
	var disabledReviewers []string

	for _, req := range reviewRequests {
		timeSinceRequest := time.Since(req.On).Round(time.Hour)
		formattedDuration := format.FormatDuration(timeSinceRequest)

		reviewer := req.From
		if req.IsTeam {
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
