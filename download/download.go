package download

import (
	"context"
	"github.com/chooban/progger/download/api"
	"github.com/chooban/progger/download/internal"
	"github.com/go-logr/logr"
)

func ListAvailableProgs(ctx context.Context) ([]api.DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)

	bContext, err := browser()

	if err != nil {
		logger.Error(err, "Could not start browser")
		return []api.DigitalComic{}, err
	}
	u := Username(ctx)
	p := Password(ctx)

	if err = internal.Login(ctx, bContext, u, p); err != nil {
		return []api.DigitalComic{}, err
	}

	if progs, err := internal.ListProgs(ctx, bContext); err != nil {
		logger.Error(err, "Could not start browser")
		return []api.DigitalComic{}, err
	} else {
		return progs, nil
	}
}
