package rules

import (
	"context"
	"path/filepath"
	"time"

	"github.com/Djiit/gong/internal/githubclient"
	"github.com/Djiit/gong/internal/ping"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Variable to allow time.Now to be mocked in tests
var timeNow = time.Now

// Rule represents a rule for matching reviewers with custom delays
type Rule struct {
	MatchName    string
	MatchTitle   string
	MatchAuthor  string // Added for matching PR authors
	Delay        int
	Enabled      bool
	Integrations []ping.Integration // List of integrations for this rule
}

// Each rule can override the global delay for specific reviewers matching the glob pattern
// or PR titles matching the glob pattern.
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
			nameMatched := false
			titleMatched := false
			authorMatched := false

			// Check if reviewer name matches the pattern, if a pattern is provided
			if rule.MatchName != "" {
				if matched, _ := filepath.Match(rule.MatchName, req.From); matched {
					nameMatched = true
				}
			}

			// Check if PR title matches the pattern, if a pattern is provided
			if rule.MatchTitle != "" && req.PRTitle != "" {
				if matched, _ := filepath.Match(rule.MatchTitle, req.PRTitle); matched {
					titleMatched = true
				}
			}

			// Check if PR author matches the pattern, if a pattern is provided
			log.Debug().Msgf("Checking if PR author %s matches pattern %s", req.PRAuthor, rule.MatchAuthor)
			if rule.MatchAuthor != "" && req.PRAuthor != "" {
				log.Debug().Msgf("Checking if PR author %s matches pattern %s", req.PRAuthor, rule.MatchAuthor)
				if matched, _ := filepath.Match(rule.MatchAuthor, req.PRAuthor); matched {
					authorMatched = true
				}
			}

			// Determine if this rule applies
			ruleApplies := false

			// All conditions must match if multiple are provided
			if rule.MatchName != "" && rule.MatchTitle != "" && rule.MatchAuthor != "" {
				ruleApplies = nameMatched && titleMatched && authorMatched
			} else if rule.MatchName != "" && rule.MatchTitle != "" {
				ruleApplies = nameMatched && titleMatched
			} else if rule.MatchName != "" && rule.MatchAuthor != "" {
				ruleApplies = nameMatched && authorMatched
			} else if rule.MatchTitle != "" && rule.MatchAuthor != "" {
				ruleApplies = titleMatched && authorMatched
			} else if rule.MatchName != "" {
				ruleApplies = nameMatched
			} else if rule.MatchTitle != "" {
				ruleApplies = titleMatched
			} else if rule.MatchAuthor != "" {
				ruleApplies = authorMatched
			}

			// Apply the rule if it matches
			if ruleApplies {
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
	rulesConfig := viper.Get("rules")
	// If rules is a slice, process each rule
	if rulesSlice, ok := rulesConfig.([]interface{}); ok {
		for _, r := range rulesSlice {
			if ruleMap, ok := r.(map[string]interface{}); ok {
				rule := Rule{}

				if matchName, ok := ruleMap["matchname"].(string); ok {
					rule.MatchName = matchName
				}

				// Parse the matchTitle field from the config
				if matchTitle, ok := ruleMap["matchtitle"].(string); ok {
					rule.MatchTitle = matchTitle
				}

				// Parse the matchAuthor field from the config
				if matchAuthor, ok := ruleMap["matchauthor"].(string); ok {
					rule.MatchAuthor = matchAuthor
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
				// Add rules with a valid match pattern (either matchName, matchTitle, or matchAuthor)
				if rule.MatchName != "" || rule.MatchTitle != "" || rule.MatchAuthor != "" {
					ruleset = append(ruleset, rule)
				}
			}
		}
	}
	log.Debug().Msgf("Parsed rules: %v", ruleset)
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
