package llm

import "context"

type ctxKey int

const talkLogSessionKey ctxKey = 1

// WithSessionID stores a talk-log session ID in the context.
func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, talkLogSessionKey, id)
}

// SessionID retrieves the talk-log session ID from the context.
// Returns empty string if not set.
func SessionID(ctx context.Context) string {
	v, _ := ctx.Value(talkLogSessionKey).(string)
	return v
}
