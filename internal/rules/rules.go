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
	MatchName   string
	Delay       int
	Enabled     bool   // Whether this rule is enabled
	Integration string // The integration to use for this rule (optional)
}

// Each rule can override the global delay for specific reviewers matching the glob pattern
// It also updates the Delay, Enabled, ShouldPing, and Integration field for each request.
func ApplyRules(ctx context.Context, requests []githubclient.ReviewRequest, rules []Rule) []ping.PingRequest {
	var pingRequests []ping.PingRequest
	now := timeNow()

	for _, req := range requests {
		pingReq := ping.PingRequest{}
		pingReq.Req = req
		pingReq.Delay = ctx.Value("delay").(int)
		pingReq.Enabled = ctx.Value("enabled").(bool)
		pingReq.Integration = ctx.Value("integration").(string)

		// Check if any rule matches this reviewer
		for _, rule := range rules {
			if matched, _ := filepath.Match(rule.MatchName, req.From); matched {
				pingReq.Delay = rule.Delay
				pingReq.Enabled = rule.Enabled

				// Override integration only if specified in the rule
				if rule.Integration != "" {
					pingReq.Integration = rule.Integration
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

					if ruleIntegration, ok := ruleMap["integration"].(string); ok {
						rule.Integration = ruleIntegration
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
