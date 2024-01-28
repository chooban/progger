package scan

import (
	"context"
	"github.com/chooban/progger/scan/env"
)

// NewContext returns a new Context, derived from ctx, which carries the
// provided config.
func NewContext(ctx context.Context, appEnv env.AppEnv) context.Context {
	return context.WithValue(ctx, contextKey{}, appEnv)
}

// contextKey is how we find configuration in a context.Context.
type contextKey struct{}

func fromContextOrDefaults(ctx context.Context) env.AppEnv {
	if v, ok := ctx.Value(contextKey{}).(env.AppEnv); ok {
		return v
	}

	return env.NewAppEnv()
}
