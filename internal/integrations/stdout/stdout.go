package stdout

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/Djiit/gong/internal/format"
	"github.com/Djiit/gong/internal/ping"
)

// DefaultTemplate is the default template used for stdout output
const DefaultTemplate = `{{- if .ActiveReviewers -}}
Pinging: {{ range $i, $r := .ActiveReviewers }}{{ if $i }}, {{ end }}{{ $r }}{{ end }}
{{- end -}}
{{ if and .ActiveReviewers .DisabledReviewers }}
{{ end }}
{{- if .DisabledReviewers -}}
Not pinging: {{ range $i, $r := .DisabledReviewers }}{{ if $i }}, {{ end }}{{ $r }}{{ end -}}
{{- end -}}
{{- if and (not .ActiveReviewers) (not .DisabledReviewers) -}}
No pending review requests.
{{- end -}}
`

// TemplateData holds the data for template rendering
type TemplateData struct {
	PingRequests      []ping.PingRequest
	ActiveReviewers   []string
	DisabledReviewers []string
}

func Run(ctx context.Context) {
	pingRequests := ctx.Value("pingRequests").([]ping.PingRequest)
	isDryRun := ctx.Value("dry-run").(bool)

	// Get template parameter from integrations config
	var templateStr string

	// First check if there's a template in the integration parameters
	if len(pingRequests) > 0 {
		for _, intg := range pingRequests[0].Integrations {
			if intg.Type == "stdout" {
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
		fmt.Println("[DRY RUN] Would output reviewer information to stdout")
		return
	}

	output, err := formatWithTemplate(pingRequests, templateStr)
	if err != nil {
		fmt.Printf("Error formatting output with template: %v\n", err)
		return
	}

	fmt.Println(output)
}

func formatWithTemplate(pingRequests []ping.PingRequest, templateStr string) (string, error) {
	if len(pingRequests) == 0 {
		return "No pending review requests.", nil
	}

	// Prepare template data
	data := prepareTemplateData(pingRequests)

	// Parse template
	tmpl, err := template.New("stdout").Parse(templateStr)
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

func prepareTemplateData(pingRequests []ping.PingRequest) TemplateData {
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
			activeReviewers = append(activeReviewers, reviewerInfo)
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
	}
}
