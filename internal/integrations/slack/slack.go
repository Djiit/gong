package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/Djiit/gong/internal/ping"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"
)

func Run(ctx context.Context) {
	pingRequests := ctx.Value("pingRequests").([]ping.PingRequest)
	repoOwner := ctx.Value("repoOwner").(string)
	repoName := ctx.Value("repoName").(string)
	prNumber := ctx.Value("pr").(string)
	isDryRun := ctx.Value("dry-run").(bool)

	if len(pingRequests) == 0 {
		return
	}

	// Extract reviewers to mention
	var reviewers []string
	for _, req := range pingRequests {
		reviewers = append(reviewers, req.Req.From)
	}

	// Find the integration parameters for Slack
	// Use parameters from the first request since all requests in the same group
	// should have the same parameters for a given integration type
	var channel string
	var webhookURL string
	for _, intg := range pingRequests[0].Integrations {
		if intg.Type == "slack" { // FIXME: why check for slack here?
			// Look for channel parameter
			if ch, ok := intg.Parameters["channel"]; ok && ch != "" {
				channel = ch
			}
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

	SendSlackMessage(channel, webhookURL, repoOwner, repoName, prNumber, reviewers, isDryRun)

}

func SendSlackMessage(channel, webhookURL, repoOwner, repoName, prNumber string, reviewers []string, isDryRun bool) {
	// Create PR URL
	prURL := fmt.Sprintf("https://github.com/%s/%s/pull/%s", repoOwner, repoName, prNumber)

	// Create message
	msgText := fmt.Sprintf("PR #%s is waiting for your review : <%s|%s/%s#%s>\n",
		prNumber, prURL, repoOwner, repoName, prNumber)

	if len(reviewers) > 0 {
		msgText += "Reviewers: " + strings.Join(reviewers, ", ")
	}

	if isDryRun {
		log.Info().Msgf("[DRY RUN] Would send Slack notification to channel %s via webhook for PR #%s in %s/%s",
			channel, prNumber, repoOwner, repoName)
		log.Info().Msgf("[DRY RUN] Message: %s", msgText)
		return
	}

	log.Debug().Msgf("Sending Slack notification to channel %s via webhook for PR #%s in %s/%s",
		channel, prNumber, repoOwner, repoName)

	// Create message attachment
	attachment := slack.Attachment{
		Color:      "#36a64f",
		Text:       msgText,
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
	message := &slack.WebhookMessage{
		Channel:     channel,
		Text:        fmt.Sprintf("Review requested on PR #%s", prNumber),
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(webhookURL, message)
	if err != nil {
		log.Error().Msgf("Error sending Slack notification: %s\n", err)
		return
	}
}
