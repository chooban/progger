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

func LoginDetails(ctx context.Context) (username, password string) {
	if u := ctx.Value(contextKeyUsername); u != nil {
		username = u.(string)
	} else {
		return
	}
	if p := ctx.Value(contextKeyPassword); p != nil {
		password = p.(string)
	} else {
		return "", ""
	}

	return
}
