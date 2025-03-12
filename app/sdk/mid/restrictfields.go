package mid

import (
	"context"
)

func RestrictFields(ctx context.Context, next HandlerFunc) Encoder {
	resp := next(ctx)

	return resp
}
