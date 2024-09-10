package mid

import (
	"context"

	"bitbucket.org/superiortechnologies/ichor/foundation/otel"
	"go.opentelemetry.io/otel/trace"
)

// Otel starts the otel tracing and stores the trace id in the context.
func Otel(ctx context.Context, tracer trace.Tracer, next HandlerFunc) Encoder {
	ctx = otel.InjectTracing(ctx, tracer)

	return next(ctx)
}
