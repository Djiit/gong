package integrations

import (
	"context"

	"github.com/Djiit/pingrequest/internal/integrations/comment"
	"github.com/Djiit/pingrequest/internal/integrations/stdout"
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
}
