package mid

import (
	"context"
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/sdk/mid"
	"bitbucket.org/superiortechnologies/ichor/foundation/logger"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
)

// Errors executes the errors middleware functionality.
func Errors(log *logger.Logger) web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.Errors(ctx, log, next)
	}

	return addMidFunc(midFunc)
}
