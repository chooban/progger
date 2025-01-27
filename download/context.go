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
	ContextKeyUsername       = contextKey("progger-username")
	ContextKeyPassword       = contextKey("progger-password")
	ContextKeyBrowserContext = contextKey("progger-browser-context")
)

func browserContextDir(ctx context.Context) (d string, err error) {
	if v := ctx.Value(ContextKeyBrowserContext); v != nil {
		d = v.(string)
	} else {
		err = errors.New("browser context not found")
	}
	return d, err
}

func loginDetails(ctx context.Context) (username, password string, err error) {
	if u := ctx.Value(ContextKeyUsername); u != nil {
		username = u.(string)
	} else {
		return "", "", errors.New("username not found")
	}
	if p := ctx.Value(ContextKeyPassword); p != nil {
		password = p.(string)
	} else {
		println("password not found")
		return "", "", errors.New("credentials not found")
	}

	return
}
