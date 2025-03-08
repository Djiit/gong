package comment

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
	"github.com/google/go-github/v69/github"
	"github.com/spf13/viper"
)

func Run(ctx context.Context) {
	pingRequests := ctx.Value("pingRequests").([]ping.PingRequest)
	repoOwner := ctx.Value("repoOwner").(string)
	repoName := ctx.Value("repoName").(string)
	prNumber := ctx.Value("pr").(string)
	isDryRun := ctx.Value("dry-run").(bool)

	// Extract the reviewers to mention from the ping requests
	var reviewers []string
	for _, req := range pingRequests {
		reviewers = append(reviewers, req.Req.From)
	}

	if isDryRun {
		fmt.Printf("[DRY RUN] Would post GitHub comment mentioning: @%s\n",
			strings.Join(reviewers, ", @"))
		return
	}

	client := githubclient.NewClient(viper.GetString("github-token"))

	prNum, err := strconv.Atoi(prNumber)
	if err != nil {
		fmt.Printf("Error converting PR number: %v\n", err)
		return
	}

	if alreadyCommented(ctx, client, repoOwner, repoName, prNum) {
		fmt.Println("Comment already exists for this PR.")
		return
	}

	if err := postComment(ctx, client, repoOwner, repoName, prNum, reviewers); err != nil {
		fmt.Printf("Error posting comment: %v\n", err)
	}
}

func alreadyCommented(ctx context.Context, client *github.Client, owner, repo string, prNumber int) bool {
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
	if err != nil {
		fmt.Printf("Error fetching comments: %v\n", err)
		return false
	}

	for _, comment := range comments {
		if strings.Contains(*comment.Body, "<!-- gong -->") {
			return true
		}
	}
	return false
}

func postComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int, reviewers []string) error {
	body := fmt.Sprintf("Awaiting reviews from: @%s\n<!-- gong -->", strings.Join(reviewers, ", @"))
	prComment := &github.IssueComment{Body: &body}
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, prComment)
	if err != nil {
		return fmt.Errorf("failed to post comment: %w", err)
	}
	return nil
}
