package integrations

import (
	"context"

	"github.com/Djiit/gong/internal/integrations/comment"
	"github.com/Djiit/gong/internal/integrations/slack"
	"github.com/Djiit/gong/internal/integrations/stdout"
)

type Integration struct {
	Name string
	Run  func(ctx context.Context)
}

var Integrations = map[string]Integration{
	"stdout": {
		Name: "Standard Output",
		Run:  stdout.Run,
	},
	"comment": {
		Name: "Comment Integration",
		Run:  comment.Run,
	},
	"slack": {
		Name: "Slack Integration",
		Run:  slack.Run,
	},
}
