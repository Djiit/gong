package ping

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Djiit/pingrequest/internal/githubclient"
	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	repository string
	repoOwner  string
	repoName   string
	pr         string
)

// pingCmd represents the ping command
var PingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping PR reviewers to remind them",
	Long:  `Ping PR reviewers to remind them to review the Pull Request.`,
	Run: func(cmd *cobra.Command, args []string) {
		repository := viper.GetString("repository")
		if repository == "" {
			log.Fatal("Repository must be specified in the format owner/repo")
		}

		pr := viper.GetString("pr")
		if pr == "" {
			log.Fatal("PR number must be specified")
		}

		repoParts := strings.Split(repository, "/")
		if len(repoParts) != 2 {
			log.Fatalf("Invalid repository format. Expected owner/repo, got %s", repository)
		}
		repoOwner = repoParts[0]
		repoName = repoParts[1]

		client := githubclient.NewClient(viper.GetString("github-token"))
		reviewers, err := getPRReviewRequests(client, repoOwner, repoName, pr)
		if err != nil {
			log.Fatalf("Error retrieving review requests: %v", err)
		}

		if len(reviewers) == 0 {
			fmt.Printf("No reviewers found for PR #%s.\n", pr)
		} else {
			fmt.Printf("Review requests for PR #%s: %v\n", pr, reviewers)
		}
	},
}

func getPRReviewRequests(client *github.Client, owner, repo string, prNumber string) ([]string, error) {
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

	return reviewers, nil
}

func init() {
	PingCmd.PersistentFlags().StringVar(&repository, "repository", repository, "Repository in the format owner/repo")
	PingCmd.PersistentFlags().StringVar(&pr, "pr", pr, "Pull Request number")
	viper.BindPFlags(PingCmd.PersistentFlags())
}
