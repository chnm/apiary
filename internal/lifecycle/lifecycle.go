// Package lifecycle coordinates server startup and graceful shutdown.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Server is the lifecycle surface required by Run.
type Server interface {
	Run() error
	Shutdown(context.Context) error
}

// Run starts server and blocks until it exits or ctx is canceled. On
// cancellation, Run waits for graceful shutdown and for the serving goroutine
// to finish before returning.
func Run(
	ctx context.Context,
	server Server,
	shutdownTimeout time.Duration,
) error {
	if shutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}

	runErr := make(chan error, 1)
	go func() {
		runErr <- server.Run()
	}()

	select {
	case err := <-runErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("gracefully shut down server: %w", err)
	}

	select {
	case err := <-runErr:
		return err
	case <-shutdownCtx.Done():
		return fmt.Errorf(
			"wait for server to stop: %w",
			shutdownCtx.Err(),
		)
	}
}
