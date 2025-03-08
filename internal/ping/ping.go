package ping

import (
	"github.com/Djiit/gong/internal/githubclient"
)

// Integration represents a single integration configuration for a ping request
type Integration struct {
	Type       string            // Type of integration (e.g., "slack", "stdout", "comment")
	Parameters map[string]string // Parameters specific to this integration instance
}

type PingRequest struct {
	Req          githubclient.ReviewRequest
	Delay        int           // The delay in seconds that applies to this reviewer
	Enabled      bool          // Whether pinging this reviewer is enabled
	ShouldPing   bool          // Whether this reviewer should be pinged (based on delay and enabled)
	Integrations []Integration // List of integrations to use for this reviewer
}
