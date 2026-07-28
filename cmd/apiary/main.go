package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	apiary "github.com/chnm/apiary"
	"github.com/chnm/apiary/internal/lifecycle"
)

const gracefulShutdownTimeout = 8 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal("error running the server: ", err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	server := apiary.NewServer(ctx)
	return lifecycle.Run(ctx, server, gracefulShutdownTimeout)
}
