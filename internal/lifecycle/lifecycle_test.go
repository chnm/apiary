package lifecycle

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeServer struct {
	runStarted       chan struct{}
	runFinished      chan struct{}
	runErr           error
	shutdownCalled   atomic.Bool
	shutdownBehavior func(context.Context) error
}

func newFakeServer() *fakeServer {
	return &fakeServer{
		runStarted:  make(chan struct{}),
		runFinished: make(chan struct{}),
	}
}

func (server *fakeServer) Run() error {
	close(server.runStarted)
	<-server.runFinished
	return server.runErr
}

func (server *fakeServer) Shutdown(ctx context.Context) error {
	server.shutdownCalled.Store(true)
	if server.shutdownBehavior != nil {
		return server.shutdownBehavior(ctx)
	}
	close(server.runFinished)
	return nil
}

func TestRunShutsDownAfterImmediateCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	server := newFakeServer()

	if err := Run(ctx, server, time.Second); err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}
	if !server.shutdownCalled.Load() {
		t.Fatal("Shutdown was not called")
	}
}

func TestRunReturnsServerErrorWithoutShutdown(t *testing.T) {
	wantErr := errors.New("listener failed")
	server := newFakeServer()
	server.runErr = wantErr
	close(server.runFinished)

	err := Run(context.Background(), server, time.Second)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Run error = %v, want %v", err, wantErr)
	}
	if server.shutdownCalled.Load() {
		t.Fatal("Shutdown was called after the server had already stopped")
	}
}

func TestRunWaitsForShutdownAndServingGoroutine(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := newFakeServer()
	shutdownRelease := make(chan struct{})
	server.shutdownBehavior = func(context.Context) error {
		<-shutdownRelease
		close(server.runFinished)
		return nil
	}

	result := make(chan error, 1)
	go func() {
		result <- Run(ctx, server, time.Second)
	}()

	<-server.runStarted
	cancel()

	select {
	case err := <-result:
		t.Fatalf("Run returned before shutdown completed: %v", err)
	case <-time.After(20 * time.Millisecond):
	}

	close(shutdownRelease)
	if err := <-result; err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}
}

func TestRunBoundsGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := newFakeServer()
	server.shutdownBehavior = func(ctx context.Context) error {
		<-ctx.Done()
		close(server.runFinished)
		return ctx.Err()
	}

	cancel()
	err := Run(ctx, server, 20*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf(
			"Run error = %v, want context.DeadlineExceeded",
			err,
		)
	}
}

func TestRunRejectsNonPositiveShutdownTimeout(t *testing.T) {
	err := Run(context.Background(), newFakeServer(), 0)
	if err == nil {
		t.Fatal("Run returned nil for a non-positive shutdown timeout")
	}
}
