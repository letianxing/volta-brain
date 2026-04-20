package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/letianxing/volta-brain/internal/app"
	"github.com/letianxing/volta-brain/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = application.Close()
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := application.Start(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done()
	os.Exit(0)
}
