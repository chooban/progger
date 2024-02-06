package download

import "context"

type contextKey string

func (c contextKey) String() string {
	return "progger context key " + string(c)
}

var (
	contextKeyUsername = contextKey("progger-username")
	contextKeyPassword = contextKey("progger-password")
)

func WithLoginDetails(ctx context.Context, username, password string) context.Context {
	ctx = context.WithValue(ctx, contextKeyUsername, username)
	ctx = context.WithValue(ctx, contextKeyPassword, password)

	return ctx
}

func Username(ctx context.Context) string {
	if u := ctx.Value(contextKeyUsername); u != nil {
		return u.(string)
	} else {
		return ""
	}
}

func Password(ctx context.Context) string {
	if u := ctx.Value(contextKeyPassword); u != nil {
		return u.(string)
	} else {
		return ""
	}
}
