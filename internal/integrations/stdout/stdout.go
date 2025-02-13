package stdout

import (
	"context"
	"fmt"
	"strings"
)

func Run(ctx context.Context) {
	reviewers := ctx.Value("reviewers").([]string)
	output := formatReviewers(reviewers)
	fmt.Println(output)
}

func formatReviewers(reviewers []string) string {
	return fmt.Sprintf("Awaiting reviews from: %s", strings.Join(reviewers, ", "))
}
