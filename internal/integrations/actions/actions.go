package actions

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Djiit/gong/internal/format"
	"github.com/Djiit/gong/internal/ping"
)

// Run executes the actions integration which writes reviewer information to GitHub Actions environment
// variables: GITHUB_OUTPUT and GITHUB_ENV.
func Run(ctx context.Context) {
	pingRequests := ctx.Value("pingRequests").([]ping.PingRequest)
	isDryRun := ctx.Value("dry-run").(bool)

	if isDryRun {
		fmt.Println("[DRY RUN] Would write reviewer information to GitHub Actions environment variables")
		return
	}

	outputFilePath := os.Getenv("GITHUB_OUTPUT")
	envFilePath := os.Getenv("GITHUB_ENV")

	if outputFilePath == "" && envFilePath == "" {
		fmt.Println("GitHub Actions environment variables not detected. This integration is meant to be used in GitHub Actions.")
		return
	}

	// Process ping requests
	enabledReviewers, disabledReviewers := processRequests(pingRequests)

	// Write to GITHUB_OUTPUT
	if outputFilePath != "" {
		writeToGitHubOutput(outputFilePath, enabledReviewers, disabledReviewers)
	}

	// Write to GITHUB_ENV
	if envFilePath != "" {
		writeToGitHubEnv(envFilePath, enabledReviewers, disabledReviewers)
	}

	// Also print the information to stdout for visibility
	outputInfo := formatOutput(enabledReviewers, disabledReviewers)
	fmt.Println("GitHub Actions Integration results:")
	fmt.Println(outputInfo)
}

// processRequests processes ping requests and separates them into enabled and disabled reviewers
func processRequests(pingRequests []ping.PingRequest) ([]string, []string) {
	var enabledReviewers []string
	var disabledReviewers []string

	for _, req := range pingRequests {
		timeSinceRequest := time.Since(req.Req.On).Round(time.Hour)
		formattedDuration := format.FormatDuration(timeSinceRequest)

		reviewer := req.Req.From
		if req.Req.IsTeam {
			reviewer += " (team)"
		}

		if req.ShouldPing {
			enabledReviewers = append(enabledReviewers, reviewer)
		} else {
			status := "waiting"
			if !req.Enabled {
				status = "disabled"
			}
			disabledReviewers = append(disabledReviewers, fmt.Sprintf("%s (%s ago, status: %s)",
				reviewer, formattedDuration, status))
		}
	}

	return enabledReviewers, disabledReviewers
}

// writeToGitHubOutput writes reviewer information to GITHUB_OUTPUT file
func writeToGitHubOutput(filePath string, enabledReviewers, disabledReviewers []string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_OUTPUT file: %w", err)
	}
	defer f.Close()

	// Write comma-separated reviewers list
	if _, err := f.WriteString(fmt.Sprintf("reviewers=%s\n", strings.Join(enabledReviewers, ","))); err != nil {
		return fmt.Errorf("failed to write reviewers to GITHUB_OUTPUT: %w", err)
	}

	// Write reviewers count
	if _, err := f.WriteString(fmt.Sprintf("reviewersCount=%d\n", len(enabledReviewers))); err != nil {
		return fmt.Errorf("failed to write reviewers count to GITHUB_OUTPUT: %w", err)
	}

	// Write multiline reviewers details
	if _, err := f.WriteString("reviewersDetails<<EOF\n"); err != nil {
		return fmt.Errorf("failed to write reviewers delimiter to GITHUB_OUTPUT: %w", err)
	}

	// List all reviewers with their status
	var allReviewers []string

	// Add enabled reviewers
	for _, reviewer := range enabledReviewers {
		allReviewers = append(allReviewers, fmt.Sprintf("%s (status: enabled)", reviewer))
	}

	// Add disabled reviewers (they already have status information)
	allReviewers = append(allReviewers, disabledReviewers...)

	if _, err := f.WriteString(strings.Join(allReviewers, "\n") + "\n"); err != nil {
		return fmt.Errorf("failed to write reviewers to GITHUB_OUTPUT: %w", err)
	}

	if _, err := f.WriteString("EOF\n"); err != nil {
		return fmt.Errorf("failed to write reviewers end delimiter to GITHUB_OUTPUT: %w", err)
	}

	return nil
}

// writeToGitHubEnv writes reviewer information to GITHUB_ENV file
func writeToGitHubEnv(filePath string, enabledReviewers, disabledReviewers []string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open GITHUB_ENV file: %w", err)
	}
	defer f.Close()

	// Write comma-separated reviewers list
	if _, err := f.WriteString(fmt.Sprintf("GONG_REVIEWERS=%s\n", strings.Join(enabledReviewers, ","))); err != nil {
		return fmt.Errorf("failed to write GONG_REVIEWERS to GITHUB_ENV: %w", err)
	}

	// Write reviewers count
	if _, err := f.WriteString(fmt.Sprintf("GONG_REVIEWERS_COUNT=%d\n", len(enabledReviewers))); err != nil {
		return fmt.Errorf("failed to write reviewers count to GITHUB_ENV: %w", err)
	}

	// Add multiline details variable
	if _, err := f.WriteString("GONG_REVIEWERS_DETAILS<<EOF\n"); err != nil {
		return fmt.Errorf("failed to write GONG_REVIEWERS_DETAILS delimiter to GITHUB_ENV: %w", err)
	}

	// List all reviewers with their status
	var allReviewers []string

	// Add enabled reviewers
	for _, reviewer := range enabledReviewers {
		allReviewers = append(allReviewers, fmt.Sprintf("%s (status: enabled)", reviewer))
	}

	// Add disabled reviewers (they already have status information)
	allReviewers = append(allReviewers, disabledReviewers...)

	if _, err := f.WriteString(strings.Join(allReviewers, "\n") + "\n"); err != nil {
		return fmt.Errorf("failed to write GONG_REVIEWERS_DETAILS to GITHUB_ENV: %w", err)
	}

	if _, err := f.WriteString("EOF\n"); err != nil {
		return fmt.Errorf("failed to write GONG_REVIEWERS_DETAILS end delimiter to GITHUB_ENV: %w", err)
	}

	return nil
}

// formatOutput creates a human-readable output of the processed ping requests
func formatOutput(enabledReviewers, disabledReviewers []string) string {
	if len(enabledReviewers) == 0 && len(disabledReviewers) == 0 {
		return "No pending review requests."
	}

	var result strings.Builder

	if len(enabledReviewers) > 0 {
		result.WriteString(fmt.Sprintf("Enabled reviewers (%d): %s\n",
			len(enabledReviewers),
			strings.Join(enabledReviewers, ", ")))
	}

	if len(disabledReviewers) > 0 {
		result.WriteString(fmt.Sprintf("Disabled/waiting reviewers (%d): %s",
			len(disabledReviewers),
			strings.Join(disabledReviewers, ", ")))
	}

	return result.String()
}
