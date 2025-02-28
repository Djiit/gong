package githubclient

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v69/github"
	"golang.org/x/oauth2"
)

func NewClient(githubToken string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}

type ReviewRequest struct {
	From string
	On   time.Time
}

func GetReviewRequests(client *github.Client, owner, repo string, prNumber string) ([]string, error) {
	ctx := context.Background()

	prNum, err := strconv.Atoi(prNumber)
	if err != nil {
		return nil, err
	}

	reviewRequests, _, err := client.PullRequests.ListReviewers(ctx, owner, repo, prNum, nil)
	if err != nil {
		return nil, err
	}

	var reviewers []string
	for _, user := range reviewRequests.Users {
		reviewers = append(reviewers, *user.Login)
	}

	for _, team := range reviewRequests.Teams {
		reviewers = append(reviewers, *team.Name)
	}

	timeline, _, err := client.Issues.ListIssueTimeline(ctx, owner, repo, prNum, nil)
	if err != nil {
		return nil, err
	}
	fmt.Println(timeline)
	// better to get from timeline because we have timestamp
	for _, event := range timeline {
		if *event.Event == "review_requested" {
			fmt.Println(event.CreatedAt.Time)
			// from can be either a user or a team, so either GetReviewer or GetRequestedTeam witll be nil
			if event.GetRequestedTeam() != nil {
				reviewers = append(reviewers, event.GetRequestedTeam().GetName())
			} else if event.GetReviewer() != nil {
				reviewers = append(reviewers, event.GetReviewer().GetLogin())
			}

			fmt.Println("Reviewer added:", reviewers)
		}
	}
	return reviewers, nil
}
