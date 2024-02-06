package download

import (
	"context"
	"github.com/go-logr/logr"
)

func ListAvailableProgs(ctx context.Context) ([]DigitalComic, error) {
	logger := logr.FromContextOrDiscard(ctx)

	bContext, err := browser()

	if err != nil {
		logger.Error(err, "Could not start browser")
		return []DigitalComic{}, err
	}
	u := Username(ctx)
	p := Password(ctx)

	if err = Login(ctx, bContext, u, p); err != nil {
		return []DigitalComic{}, err
	}

	if progs, err := ListProgs(ctx, bContext); err != nil {
		logger.Error(err, "Could not start browser")
		return []DigitalComic{}, err
	} else {
		return progs, nil
	}
}
