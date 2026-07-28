package apiary

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestServerShutdownWaitsForActiveRequests(t *testing.T) {
	handlerStarted := make(chan struct{})
	releaseHandler := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(handlerStarted)
		<-releaseHandler
		w.WriteHeader(http.StatusNoContent)
	})

	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)

	pool, err := pgxpool.New(
		context.Background(),
		"postgres://apiary:apiary@127.0.0.1:1/apiary",
	)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	t.Cleanup(pool.Close)

	server := &Server{
		Server: httpServer.Config,
		DB:     pool,
	}

	requestDone := make(chan error, 1)
	go func() {
		response, err := http.Get(httpServer.URL)
		if err == nil {
			defer response.Body.Close()
			if response.StatusCode != http.StatusNoContent {
				err = fmt.Errorf(
					"response status = %d, want %d",
					response.StatusCode,
					http.StatusNoContent,
				)
			}
		}
		requestDone <- err
	}()

	<-handlerStarted
	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- server.Shutdown(context.Background())
	}()

	returnedEarly := false
	select {
	case err := <-shutdownDone:
		returnedEarly = true
		if err != nil {
			t.Errorf("Shutdown returned an error: %v", err)
		}
	case <-time.After(20 * time.Millisecond):
	}

	close(releaseHandler)
	if err := <-requestDone; err != nil {
		t.Fatalf("active request failed during shutdown: %v", err)
	}
	if returnedEarly {
		t.Fatal("Shutdown returned before the active request completed")
	}
	if err := <-shutdownDone; err != nil {
		t.Fatalf("Shutdown returned an error: %v", err)
	}
}
