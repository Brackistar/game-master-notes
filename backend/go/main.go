package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/Brackistar/game-master-notes/backend/go/src/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		log.Fatalf("backend failed: %v", err)
	}
}
