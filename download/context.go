package download

import (
	"context"
	"errors"
)

type contextKey string

func (c contextKey) String() string {
	return "progger context key " + string(c)
}

var (
	contextKeyUsername       = contextKey("progger-username")
	contextKeyPassword       = contextKey("progger-password")
	contextKeyBrowserContext = contextKey("progger-browser-context")
)

func WithLoginDetails(ctx context.Context, username, password string) context.Context {
	ctx = context.WithValue(ctx, contextKeyUsername, username)
	ctx = context.WithValue(ctx, contextKeyPassword, password)

	return ctx
}

func WithBrowserContextDir(ctx context.Context, dir string) context.Context {
	ctx = context.WithValue(ctx, contextKeyBrowserContext, dir)

	return ctx
}

func BrowserContextDir(ctx context.Context) (d string, err error) {
	if v := ctx.Value(contextKeyBrowserContext); v != nil {
		d = v.(string)
	} else {
		err = errors.New("browser context not found")
	}
	return d, err
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
