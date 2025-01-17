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

func WithLoginDetails(parent context.Context, username, password string) context.Context {
	child := context.WithValue(parent, contextKeyUsername, username)
	child = context.WithValue(child, contextKeyPassword, password)

	return child
}

func WithBrowserContextDir(ctx context.Context, dir string) context.Context {
	return context.WithValue(ctx, contextKeyBrowserContext, dir)
}

func BrowserContextDir(ctx context.Context) (d string, err error) {
	if v := ctx.Value(contextKeyBrowserContext); v != nil {
		d = v.(string)
	} else {
		err = errors.New("browser context not found")
	}
	return d, err
}

func LoginDetails(ctx context.Context) (username, password string, err error) {
	if u := ctx.Value(contextKeyUsername); u != nil {
		username = u.(string)
	} else {
		return "", "", errors.New("username not found")
	}
	if p := ctx.Value(contextKeyPassword); p != nil {
		password = p.(string)
	} else {
		println("password not found")
		return "", "", errors.New("credentials not found")
	}

	return
}
