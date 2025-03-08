package ping

import (
	"github.com/Djiit/gong/internal/githubclient"
)

type PingRequest struct {
	Req         githubclient.ReviewRequest
	Delay       int    // The delay in seconds that applies to this reviewer
	Enabled     bool   // Whether pinging this reviewer is enabled
	ShouldPing  bool   // Whether this reviewer should be pinged (based on delay and enabled)
	Integration string // The integration to use for this reviewer
}
