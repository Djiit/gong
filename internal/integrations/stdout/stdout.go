package stdout

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

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

	// Use the shared template data preparation with full info
	data := format.PrepareTemplateData(pingRequests, "", "", "", "", true)

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
