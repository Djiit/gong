package comment

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"text/template"

	"github.com/Djiit/gong/internal/format"
	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
	"github.com/google/go-github/v69/github"
	"github.com/spf13/viper"
)

// DefaultTemplate is the default template used for comment output
const DefaultTemplate = `Awaiting reviews from: {{ range $i, $r := .ActiveReviewers }}{{ if $i }}, {{ end }}@{{ $r }}{{ end }}
<!-- gong -->`

func Run(ctx context.Context) {
	pingRequests := ctx.Value("pingRequests").([]ping.PingRequest)
	repoOwner := ctx.Value("repoOwner").(string)
	repoName := ctx.Value("repoName").(string)
	prNumber := ctx.Value("pr").(string)
	isDryRun := ctx.Value("dry-run").(bool)

	// Get template parameter from integrations config
	var templateStr string

	// First check if there's a template in the integration parameters
	if len(pingRequests) > 0 {
		for _, intg := range pingRequests[0].Integrations {
			if intg.Type == "comment" {
				// Look for template parameter
				if tmpl, ok := intg.Parameters["template"]; ok && tmpl != "" {
					templateStr = tmpl
					break
				}
			}
		}
	}

	// If no template found in integration parameters, try context
	if templateStr == "" {
		if val, ok := ctx.Value("template").(string); ok && val != "" {
			templateStr = val
		} else {
			templateStr = DefaultTemplate
		}
	}

	if isDryRun {
		output, err := formatWithTemplate(pingRequests, templateStr)
		if err != nil {
			fmt.Printf("Error formatting output with template: %v\n", err)
			return
		}
		fmt.Printf("[DRY RUN] Would post GitHub comment:\n%s\n", output)
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

	output, err := formatWithTemplate(pingRequests, templateStr)
	if err != nil {
		fmt.Printf("Error formatting comment with template: %v\n", err)
		return
	}

	if err := postComment(ctx, client, repoOwner, repoName, prNum, output); err != nil {
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
		if strings.Contains(comment.GetBody(), "<!-- gong -->") {
			return true
		}
	}

	return false
}

func postComment(ctx context.Context, client *github.Client, owner, repo string, prNumber int, body string) error {
	comment := &github.IssueComment{Body: github.String(body)}
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, prNumber, comment)
	return err
}

func formatWithTemplate(pingRequests []ping.PingRequest, templateStr string) (string, error) {
	if len(pingRequests) == 0 {
		return "No pending review requests.\n<!-- gong -->", nil
	}

	// Use the shared template data preparation
	data := format.PrepareTemplateData(pingRequests, "", "", "", "", false)

	// Parse template
	tmpl, err := template.New("comment").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("template parsing error: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}
