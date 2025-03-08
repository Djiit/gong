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
	repository string
	repoOwner  string
	repoName   string
	pr         string
	delay      int
	enabled    bool
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

		repoParts := strings.Split(repository, "/")
		if len(repoParts) != 2 {
			log.Fatalf("Invalid repository format. Expected owner/repo, got %s", repository)
		}
		repoOwner = repoParts[0]
		repoName = repoParts[1]

		// Create context with all necessary values
		ctx := context.WithValue(cmd.Context(), "dry-run", isDryRun)
		ctx = context.WithValue(ctx, "enabled", enabled)
		ctx = context.WithValue(ctx, "delay", delay)
		ctx = context.WithValue(ctx, "repoOwner", repoOwner)
		ctx = context.WithValue(ctx, "repoName", repoName)
		ctx = context.WithValue(ctx, "pr", pr)

		// Parse global integrations from config
		globalIntegrations := rules.ParseGlobalIntegrations()

		// If no integrations are configured, add default stdout
		if len(globalIntegrations) == 0 {
			globalIntegrations = []ping.Integration{
				{
					Type:       "stdout",
					Parameters: make(map[string]string),
				},
			}
		}

		ctx = context.WithValue(ctx, "integrations", globalIntegrations)

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

		// Store all ping requests in context
		ctx = context.WithValue(ctx, "pingRequests", pingRequests)

		// Group ping requests by integration type
		integrationGroups := make(map[string][]ping.PingRequest)

		// For each ping request that should be pinged, process each of its integrations
		for _, req := range pingRequests {
			if !req.ShouldPing {
				continue
			}

			// Process each integration for this request
			for _, integration := range req.Integrations {
				integrationGroups[integration.Type] = append(integrationGroups[integration.Type], req)
			}
		}

		// Process each integration group separately
		for integrationType, requests := range integrationGroups {
			integrationFunc, ok := integrations.Integrations[integrationType]
			if !ok {
				log.Printf("Warning: Unknown integration: %s, skipping associated reviewers", integrationType)
				continue
			}

			// Create a new context with just the requests for this integration
			integrationCtx := context.WithValue(ctx, "pingRequests", requests)

			// Execute the integration
			integrationFunc.Run(integrationCtx)
		}
	},
}

func init() {
	PingCmd.PersistentFlags().StringVarP(&repository, "repository", "r", repository, "Repository in the format owner/repo (auto-detected if not specified)")
	PingCmd.PersistentFlags().StringVar(&pr, "pr", pr, "Pull Request number")
	PingCmd.PersistentFlags().IntVarP(&delay, "delay", "d", 0, "Delay in seconds before pinging reviewers (default: 0, ping immediately)")
	PingCmd.PersistentFlags().BoolVar(&enabled, "enabled", true, "Enable or disable pinging functionality (default: true)")
	err := viper.BindPFlags(PingCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("Error binding flags: %v", err)
	}
	viper.SetDefault("delay", 0)
	viper.SetDefault("enabled", true)
}
