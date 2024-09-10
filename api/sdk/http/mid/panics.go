package mid

import (
	"context"
	"net/http"

	"bitbucket.org/superiortechnologies/ichor/app/sdk/mid"
	"bitbucket.org/superiortechnologies/ichor/foundation/web"
)

// Panics executes the panic middleware functionality.
func Panics() web.MidFunc {
	midFunc := func(ctx context.Context, r *http.Request, next mid.HandlerFunc) mid.Encoder {
		return mid.Panics(ctx, next)
	}

	return addMidFunc(midFunc)
}
