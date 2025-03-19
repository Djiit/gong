package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/Djiit/gong/internal/ping"
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
		if intg.Type == "slack" {
			// Look for channel parameter
			if ch, ok := intg.Parameters["channel"]; ok && ch != "" {
				channel = ch
			}
			// Check if webhook override exists in this integration
			if wh, ok := intg.Parameters["webhook"]; ok && wh != "" {
				webhookURL = wh
			}
		}
	}

	// If no channel found, use default
	if channel == "" {
		channel = "general" // Default channel if not specified
	}

	// If no webhook URL found in integration parameters, use global config
	if webhookURL == "" {
		webhookURL = viper.GetString("slack-webhook")
		// If still not found, check the integrations section of config
		if webhookURL == "" && viper.IsSet("integrations") {
			// Try to find slack integration in global config
			integrations := viper.Get("integrations")
			if integrationsSlice, ok := integrations.([]interface{}); ok {
				for _, intg := range integrationsSlice {
					if intgMap, ok := intg.(map[string]interface{}); ok {
						if intgType, ok := intgMap["type"].(string); ok && intgType == "slack" {
							if params, ok := intgMap["params"].(map[string]interface{}); ok {
								if wh, ok := params["webhook"].(string); ok && wh != "" {
									webhookURL = wh
									break
								}
							}
						}
					}
				}
			}
		}
	}

	// Fail if no webhook URL is found
	if webhookURL == "" {
		fmt.Printf("Error: No Slack webhook URL found in configuration. Skipping Slack notifications.\n")
		return
	}

	// Create PR URL
	prURL := fmt.Sprintf("https://github.com/%s/%s/pull/%s", repoOwner, repoName, prNumber)

	// Create message
	msgText := fmt.Sprintf("PR #%s is waiting for your review : <%s|%s/%s#%s>\n",
		prNumber, prURL, repoOwner, repoName, prNumber)

	if len(reviewers) > 0 {
		// Format reviewers with @ to mention them
		formattedReviewers := make([]string, len(reviewers))
		for i, reviewer := range reviewers {
			formattedReviewers[i] = "@" + reviewer
		}
		msgText += "Reviewers: " + strings.Join(formattedReviewers, ", ")
	}

	if isDryRun {
		fmt.Printf("[DRY RUN] Would send Slack notification to channel %s via webhook for PR #%s in %s/%s mentioning: @%s\n",
			channel, prNumber, repoOwner, repoName, strings.Join(reviewers, ", @"))
		fmt.Printf("[DRY RUN] Message: %s\n", msgText)
		return
	}

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
		Text:        fmt.Sprintf("Les reviewers suivants sont attendus sur la PR #%s", prNumber),
		Attachments: []slack.Attachment{attachment},
	}

	err := slack.PostWebhook(webhookURL, message)
	if err != nil {
		fmt.Printf("Error sending Slack notification: %s\n", err)
		return
	}

	fmt.Printf("Sent Slack notification to channel %s for PR #%s in %s/%s mentioning: @%s\n",
		channel, prNumber, repoOwner, repoName, strings.Join(reviewers, ", @"))
}
