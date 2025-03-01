package stdout

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Djiit/pingrequest/internal/format"
	"github.com/Djiit/pingrequest/internal/githubclient"
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

	var reviewers []string
	for _, req := range reviewRequests {
		timeSinceRequest := time.Since(req.On).Round(time.Hour)
		formattedDuration := format.FormatDuration(timeSinceRequest)

		reviewer := req.From
		if req.IsTeam {
			reviewer += " (team)"
		}

		reviewers = append(reviewers, fmt.Sprintf("%s (%s ago)", reviewer, formattedDuration))
	}

	return fmt.Sprintf("Awaiting reviews from: %s", strings.Join(reviewers, ", "))
}
