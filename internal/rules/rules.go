package rules

import (
	"context"
	"path/filepath"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
	"github.com/spf13/viper"
)

// Variable to allow time.Now to be mocked in tests
var timeNow = time.Now

// Rule represents a rule for matching reviewers with custom delays
type Rule struct {
	MatchName    string
	Delay        int
	Enabled      bool
	Integrations []ping.Integration // List of integrations for this rule
}

// Each rule can override the global delay for specific reviewers matching the glob pattern
// It also updates the Delay, Enabled, ShouldPing, and Integrations field for each request.
func ApplyRules(ctx context.Context, requests []githubclient.ReviewRequest, rules []Rule) []ping.PingRequest {
	var pingRequests []ping.PingRequest
	now := timeNow()

	// Get global integrations from context
	var globalIntegrations []ping.Integration
	if intgs, ok := ctx.Value("integrations").([]ping.Integration); ok {
		globalIntegrations = intgs
	}

	for _, req := range requests {
		pingReq := ping.PingRequest{
			Req:          req,
			Delay:        ctx.Value("delay").(int),
			Enabled:      ctx.Value("enabled").(bool),
			Integrations: make([]ping.Integration, len(globalIntegrations)),
		}

		// Copy global integrations
		copy(pingReq.Integrations, globalIntegrations)

		// Check if any rule matches this reviewer
		for _, rule := range rules {
			if matched, _ := filepath.Match(rule.MatchName, req.From); matched {
				pingReq.Delay = rule.Delay
				pingReq.Enabled = rule.Enabled

				// Override integrations if specified in the rule
				if len(rule.Integrations) > 0 {
					pingReq.Integrations = rule.Integrations
				}
				break
			}
		}

		// Determine if we should ping based on delay and enabled status
		pingReq.ShouldPing = pingReq.Enabled && (pingReq.Delay <= 0 || now.Sub(req.On).Seconds() >= float64(pingReq.Delay))
		pingRequests = append(pingRequests, pingReq)
	}

	return pingRequests
}

// ParseIntegration parses a single integration configuration from viper
func ParseIntegration(intgMap map[string]interface{}) ping.Integration {
	integration := ping.Integration{
		Parameters: make(map[string]string),
	}

	// Extract integration type
	if integrationType, ok := intgMap["type"].(string); ok {
		integration.Type = integrationType
	}

	// Extract parameters if they exist
	if params, ok := intgMap["params"].(map[string]interface{}); ok {
		for k, v := range params {
			if strValue, ok := v.(string); ok {
				integration.Parameters[k] = strValue
			}
		}
	}

	return integration
}

// ParseRules extracts rules configuration from viper
func ParseRules() []Rule {
	var ruleset []Rule
	if viper.IsSet("rules") {
		rulesConfig := viper.Get("rules")

		// If rules is a slice, process each rule
		if rulesSlice, ok := rulesConfig.([]interface{}); ok {
			for _, r := range rulesSlice {
				if ruleMap, ok := r.(map[string]interface{}); ok {
					rule := Rule{}

					if matchName, ok := ruleMap["matchName"].(string); ok {
						rule.MatchName = matchName
					}

					if delay, ok := ruleMap["delay"].(int); ok {
						rule.Delay = delay
					}

					if enabled, ok := ruleMap["enabled"].(bool); ok {
						rule.Enabled = enabled
					}

					// Extract integrations if they exist
					if integrations, ok := ruleMap["integrations"].([]interface{}); ok {
						for _, intg := range integrations {
							if intgMap, ok := intg.(map[string]interface{}); ok {
								integration := ParseIntegration(intgMap)
								if integration.Type != "" {
									rule.Integrations = append(rule.Integrations, integration)
								}
							}
						}
					}

					if rule.MatchName != "" { // Only add rules with a valid match pattern
						ruleset = append(ruleset, rule)
					}
				}
			}
		}
	}
	return ruleset
}

// ParseGlobalIntegrations extracts global integration configurations from viper
func ParseGlobalIntegrations() []ping.Integration {
	var integrations []ping.Integration

	if viper.IsSet("integrations") {
		integrationsConfig := viper.Get("integrations")

		if integrationsSlice, ok := integrationsConfig.([]interface{}); ok {
			for _, intg := range integrationsSlice {
				if intgMap, ok := intg.(map[string]interface{}); ok {
					integration := ParseIntegration(intgMap)
					if integration.Type != "" {
						integrations = append(integrations, integration)
					}
				}
			}
		}
	}

	return integrations
}
