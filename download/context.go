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
