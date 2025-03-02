package ping

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Djiit/pingrequest/internal/githubclient"
	"github.com/Djiit/pingrequest/internal/integrations"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	repository  string
	repoOwner   string
	repoName    string
	pr          string
	integration string
	delay       int
)

// pingCmd represents the ping command
var PingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping PR reviewers to remind them",
	Long:  `Ping PR reviewers to remind them to review the Pull Request.`,
	Run: func(cmd *cobra.Command, args []string) {
		isDryRun := viper.GetBool("dry-run")

		repository := viper.GetString("repository")
		// If repository is not specified, try to detect it
		if repository == "" {
			detectedRepo, err := githubclient.GetCurrentRepository()
			if err != nil {
				log.Fatalf("Error detecting current repository: %v. Please specify a repository using the --repository flag.", err)
			}
			if detectedRepo == "" {
				log.Fatal("Could not detect current repository. Please specify a repository using the --repository flag.")
			}
			repository = detectedRepo
			fmt.Printf("Using detected repository: %s\n", repository)
		}

		pr := viper.GetString("pr")
		if pr == "" {
			log.Fatal("PR number must be specified")
		}

		integration := viper.GetString("integration")
		if integration == "" {
			log.Fatal("Integration must be specified")
		}

		repoParts := strings.Split(repository, "/")
		if len(repoParts) != 2 {
			log.Fatalf("Invalid repository format. Expected owner/repo, got %s", repository)
		}
		repoOwner = repoParts[0]
		repoName = repoParts[1]

		client := githubclient.NewClient(viper.GetString("github-token"))
		reviewRequests, err := githubclient.GetReviewRequests(client, repoOwner, repoName, pr)
		if err != nil {
			log.Fatalf("Error retrieving review requests: %v", err)
		}

		if len(reviewRequests) == 0 {
			fmt.Printf("No reviewers found for PR #%s.\n", pr)
			return
		}

		// Apply delay filter
		delaySeconds := viper.GetInt("delay")
		if delaySeconds > 0 {
			filteredRequests := githubclient.FilterReviewRequestsByDelay(reviewRequests, delaySeconds)

			if len(filteredRequests) == 0 {
				fmt.Printf("No reviewers found for PR #%s that exceed the delay of %d seconds.\n", pr, delaySeconds)
				return
			}

			reviewRequests = filteredRequests
		}

		ctx := context.WithValue(cmd.Context(), "dry-run", isDryRun)
		ctx = context.WithValue(ctx, "reviewRequests", reviewRequests)
		ctx = context.WithValue(ctx, "repoOwner", repoOwner)
		ctx = context.WithValue(ctx, "repoName", repoName)
		ctx = context.WithValue(ctx, "pr", pr)

		integrationFunc, ok := integrations.Integrations[integration]
		if !ok {
			log.Fatalf("Unknown integration: %s", integration)
		}

		integrationFunc.Run(ctx)
	},
}

func init() {
	PingCmd.PersistentFlags().StringVarP(&repository, "repository", "r", repository, "Repository in the format owner/repo (auto-detected if not specified)")
	PingCmd.PersistentFlags().StringVar(&pr, "pr", pr, "Pull Request number")
	PingCmd.PersistentFlags().StringVarP(&integration, "integration", "i", integration, "Integration to use for pinging reviewers (e.g., stdout, comment)")
	PingCmd.PersistentFlags().IntVarP(&delay, "delay", "d", 0, "Delay in seconds before pinging reviewers (default: 0, ping immediately)")
	viper.BindPFlags(PingCmd.PersistentFlags())
	viper.SetDefault("integration", "stdout")
	viper.SetDefault("delay", 0)
}
