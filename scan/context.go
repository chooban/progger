package scan

import (
	"context"
)

// NewContext returns a new Context, derived from ctx, which carries the
// provided config.
func NewContext(ctx context.Context, appEnv AppEnv) context.Context {
	return context.WithValue(ctx, contextKey{}, appEnv)
}

// contextKey is how we find configuration in a context.Context.
type contextKey struct{}

func fromContextOrDefaults(ctx context.Context) AppEnv {
	if v, ok := ctx.Value(contextKey{}).(AppEnv); ok {
		return v
	}

	return NewAppEnv()
}
