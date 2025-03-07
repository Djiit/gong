package githubclient

import (
	"context"
	"strconv"
	"time"

	"github.com/cli/go-gh/v2/pkg/repository"
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
	From       string
	On         time.Time
	IsTeam     bool
	Delay      int  // The delay in seconds that applies to this reviewer
	Enabled    bool // Whether pinging this reviewer is enabled
	ShouldPing bool // Whether this reviewer should be pinged (based on delay and enabled)
}

// GetCurrentRepository detects the current repository from the local git context
// Returns the repository in the format "owner/repo" or an empty string and error if detection fails
func GetCurrentRepository() (string, error) {
	// Use go-gh to get the current repository
	repoInfo, err := repository.Current()
	if err != nil {
		return "", err
	}

	// Return in the format "owner/repo"
	return repoInfo.Owner + "/" + repoInfo.Name, nil
}

func GetReviewRequests(client *github.Client, owner, repo string, prNumber string) ([]ReviewRequest, error) {
	ctx := context.Background()

	prNum, err := strconv.Atoi(prNumber)
	if err != nil {
		return nil, err
	}

	reviewRequests, _, err := client.PullRequests.ListReviewers(ctx, owner, repo, prNum, nil)
	if err != nil {
		return nil, err
	}

	// Create a map to store reviewer logins and their request times
	reviewerTimestamps := make(map[string]time.Time)
	teamTimestamps := make(map[string]time.Time)

	// Fetch the timeline to get the request timestamps
	timeline, _, err := client.Issues.ListIssueTimeline(ctx, owner, repo, prNum, nil)
	if err != nil {
		return nil, err
	}

	// Process timeline events to extract timestamps
	for _, event := range timeline {
		if *event.Event == "review_requested" {
			if event.GetRequestedTeam() != nil {
				teamName := event.GetRequestedTeam().GetName()
				teamTimestamps[teamName] = event.GetCreatedAt().Time
			} else if event.GetReviewer() != nil {
				login := event.GetReviewer().GetLogin()
				reviewerTimestamps[login] = event.GetCreatedAt().Time
			}
		}
	}

	var reviewRequestsArray []ReviewRequest

	// Add individual users with their timestamps
	for _, user := range reviewRequests.Users {
		login := *user.Login
		timestamp, exists := reviewerTimestamps[login]
		if !exists {
			// If we couldn't find a timestamp, use current time as fallback
			timestamp = time.Now()
		}
		reviewRequestsArray = append(reviewRequestsArray, ReviewRequest{
			From:       login,
			On:         timestamp,
			IsTeam:     false,
			Enabled:    true,  // Default to enabled
			Delay:      0,     // Default delay will be set later
			ShouldPing: false, // Will be evaluated later
		})
	}

	// Add teams with their timestamps
	for _, team := range reviewRequests.Teams {
		teamName := *team.Name
		timestamp, exists := teamTimestamps[teamName]
		if !exists {
			// If we couldn't find a timestamp, use current time as fallback
			timestamp = time.Now()
		}
		reviewRequestsArray = append(reviewRequestsArray, ReviewRequest{
			From:       teamName,
			On:         timestamp,
			IsTeam:     true,
			Enabled:    true,  // Default to enabled
			Delay:      0,     // Default delay will be set later
			ShouldPing: false, // Will be evaluated later
		})
	}

	return reviewRequestsArray, nil
}
