package app

import (
	"context"

	"github.com/imsab23/platform-be/pkg/lifecycle"
)

func RunServer() error {
	ctx, stop := lifecycle.Signal(context.Background())
	defer stop()

	srv, err := NewServer(ctx)
	if err != nil {
		return err
	}

	err = srv.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}
