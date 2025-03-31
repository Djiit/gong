package slack

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/Djiit/gong/internal/format"
	"github.com/Djiit/gong/internal/ping"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

// DefaultTemplate is the default template used for Slack output
const DefaultTemplate = `PR #{{.PRNumber}} is waiting for review: <{{.PRURL}}|{{.RepoOwner}}/{{.RepoName}}#{{.PRNumber}}>
{{ if .ActiveReviewers }}Reviewers: {{ range $i, $r := .ActiveReviewers }}{{ if $i }}, {{ end }}{{ $r }}{{ end }}{{ end }}`

// TemplateData holds the data for template rendering
type TemplateData struct {
	PingRequests      []ping.PingRequest
	ActiveReviewers   []string
	DisabledReviewers []string
	PRNumber          string
	RepoOwner         string
	RepoName          string
	PRURL             string
}

func Run(ctx context.Context) {
	pingRequests := ctx.Value("pingRequests").([]ping.PingRequest)
	repoOwner := ctx.Value("repoOwner").(string)
	repoName := ctx.Value("repoName").(string)
	prNumber := ctx.Value("pr").(string)
	isDryRun := ctx.Value("dry-run").(bool)

	if len(pingRequests) == 0 {
		return
	}

	// Get template parameter from integrations config
	var templateStr string
	var channel string
	var webhookURL string

	// First check if there's a template in the integration parameters
	if len(pingRequests) > 0 {
		for _, intg := range pingRequests[0].Integrations {
			if intg.Type == "slack" {
				// Look for template parameter
				if tmpl, ok := intg.Parameters["template"]; ok && tmpl != "" {
					templateStr = tmpl
				}
				// Look for channel parameter
				if ch, ok := intg.Parameters["channel"]; ok && ch != "" {
					channel = ch
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

	// If no channel found, use default
	if channel == "" {
		channel = "general" // Default channel if not specified
	}

	webhookURL = viper.GetString("slack-webhook")

	// Fail if no webhook URL is found
	if webhookURL == "" {
		log.Error().Msg("No Slack webhook URL found in configuration. Skipping Slack notifications.")
		return
	}

	// Create PR URL
	prURL := fmt.Sprintf("https://github.com/%s/%s/pull/%s", repoOwner, repoName, prNumber)

	message, err := formatWithTemplate(pingRequests, templateStr, repoOwner, repoName, prNumber, prURL)
	if err != nil {
		log.Error().Err(err).Msg("Error formatting Slack message with template")
		return
	}

	if isDryRun {
		log.Info().Msgf("[DRY RUN] Would send Slack notification to channel %s via webhook for PR #%s in %s/%s",
			channel, prNumber, repoOwner, repoName)
		log.Info().Msgf("[DRY RUN] Message: %s", message)
		return
	}

	sendSlackMessage(channel, webhookURL, prURL, prNumber, message)
}

func formatWithTemplate(pingRequests []ping.PingRequest, templateStr, repoOwner, repoName, prNumber, prURL string) (string, error) {
	if len(pingRequests) == 0 {
		return "No pending review requests.", nil
	}

	// Prepare template data
	data := prepareTemplateData(pingRequests, repoOwner, repoName, prNumber, prURL)

	// Parse template
	tmpl, err := template.New("slack").Parse(templateStr)
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

func prepareTemplateData(pingRequests []ping.PingRequest, repoOwner, repoName, prNumber, prURL string) TemplateData {
	var activeReviewers []string
	var disabledReviewers []string

	for _, req := range pingRequests {
		timeSinceRequest := time.Since(req.Req.On).Round(time.Hour)
		formattedDuration := format.FormatDuration(timeSinceRequest)

		reviewer := req.Req.From
		if req.Req.IsTeam {
			reviewer += " (team)"
		}

		reviewerInfo := fmt.Sprintf("%s (%s ago, delay: %ds)",
			reviewer, formattedDuration, req.Delay)

		if req.ShouldPing {
			activeReviewers = append(activeReviewers, reviewer) // Just the name for mentions
		} else {
			status := "waiting"
			if !req.Enabled {
				status = "disabled"
			}
			disabledReviewers = append(disabledReviewers, fmt.Sprintf("%s, status: %s", reviewerInfo, status))
		}
	}

	return TemplateData{
		PingRequests:      pingRequests,
		ActiveReviewers:   activeReviewers,
		DisabledReviewers: disabledReviewers,
		PRNumber:          prNumber,
		RepoOwner:         repoOwner,
		RepoName:          repoName,
		PRURL:             prURL,
	}
}

func sendSlackMessage(channel, webhookURL, prURL, prNumber, message string) {
	log.Debug().Msgf("Sending Slack notification to channel %s via webhook", channel)

	// Create message attachment
	attachment := slack.Attachment{
		Color:      "#36a64f",
		Text:       message,
		AuthorName: "🛎️ gong",
		AuthorLink: "https://github.com/Djiit/gong",
		Footer:     "Sent via gong",
		FooterIcon: "https://github.com/favicon.ico",
		Actions: []slack.AttachmentAction{
			{
				Type:  "button",
				Text:  "Voir la PR",
				URL:   prURL,
				Style: "primary",
			},
		},
	}

	// Send to Slack using webhook
	slackMessage := &slack.WebhookMessage{
		Channel:     channel,
		Text:        fmt.Sprintf("Review requested on PR #%s", prNumber),
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(webhookURL, slackMessage)
	if err != nil {
		log.Error().Err(err).Msg("Error sending Slack notification")
		return
	}
}
