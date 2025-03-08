package ping

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/integrations"
	"github.com/Djiit/gong/internal/ping"
	"github.com/Djiit/gong/internal/rules"
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
	enabled     bool
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

		ctx := context.WithValue(cmd.Context(), "dry-run", isDryRun)
		ctx = context.WithValue(ctx, "enabled", enabled)
		ctx = context.WithValue(ctx, "delay", delay)
		ctx = context.WithValue(ctx, "repoOwner", repoOwner)
		ctx = context.WithValue(ctx, "repoName", repoName)
		ctx = context.WithValue(ctx, "pr", pr)
		ctx = context.WithValue(ctx, "integration", integration)

		// Get review requests
		client := githubclient.NewClient(viper.GetString("github-token"))
		reviewRequests, err := githubclient.GetReviewRequests(client, repoOwner, repoName, pr)
		if err != nil {
			log.Fatalf("Error retrieving review requests: %v", err)
		}

		if len(reviewRequests) == 0 {
			fmt.Printf("No reviewers found for PR #%s.\n", pr)
			return
		}

		// Parse rules from config
		ruleset := rules.ParseRules()

		// Enrich review requests data with rules
		pingRequests := rules.ApplyRules(ctx, reviewRequests, ruleset)

		ctx = context.WithValue(ctx, "pingRequests", pingRequests)

		// Group review requests by integration
		integrationGroups := make(map[string][]ping.PingRequest)
		for _, req := range pingRequests {
			if req.ShouldPing {
				integrationName := req.Integration
				// If no integration specified for this reviewer, use the default
				if integrationName == "" {
					integrationName = integration
				}
				integrationGroups[integrationName] = append(integrationGroups[integrationName], req)
			}
		}

		// Process each integration group separately
		for integrationName, requests := range integrationGroups {
			integrationFunc, ok := integrations.Integrations[integrationName]
			if !ok {
				log.Printf("Warning: Unknown integration: %s, skipping associated reviewers", integrationName)
				continue
			}

			// Create a new context with just the requests for this integration
			integrationCtx := context.WithValue(ctx, "reviewRequests", requests)
			integrationFunc.Run(integrationCtx)
		}
	},
}

func init() {
	PingCmd.PersistentFlags().StringVarP(&repository, "repository", "r", repository, "Repository in the format owner/repo (auto-detected if not specified)")
	PingCmd.PersistentFlags().StringVar(&pr, "pr", pr, "Pull Request number")
	PingCmd.PersistentFlags().StringVarP(&integration, "integration", "i", integration, "Integration to use for pinging reviewers (e.g., stdout, comment)")
	PingCmd.PersistentFlags().IntVarP(&delay, "delay", "d", 0, "Delay in seconds before pinging reviewers (default: 0, ping immediately)")
	PingCmd.PersistentFlags().BoolVar(&enabled, "enabled", true, "Enable or disable pinging functionality (default: true)")
	err := viper.BindPFlags(PingCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("Error binding flags: %v", err)
	}
	viper.SetDefault("integration", "stdout")
	viper.SetDefault("delay", 0)
	viper.SetDefault("enabled", true)
}
