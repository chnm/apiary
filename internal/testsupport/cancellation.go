// Package testsupport contains reusable helpers for Apiary tests.
package testsupport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type requestContextKey struct{}

const requestContextMarker = "dataset-request-context"

// CancellationCase describes a handler request whose database work should stop
// when the request context is canceled.
type CancellationCase struct {
	Name      string
	Path      string
	RouteVars map[string]string
	Handler   func(*pgxpool.Pool) http.HandlerFunc
}

// TestRequestCancellation runs cancellation propagation checks for handlers.
func TestRequestCancellation(t *testing.T, tests []CancellationCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			observedContext := make(chan context.Context, 1)
			config, err := pgxpool.ParseConfig("postgres://apiary:apiary@127.0.0.1:1/apiary")
			if err != nil {
				t.Fatalf("parse database config: %v", err)
			}
			config.BeforeConnect = func(ctx context.Context, _ *pgx.ConnConfig) error {
				observedContext <- ctx
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(500 * time.Millisecond):
					return errors.New("database context was not canceled")
				}
			}

			pool, err := pgxpool.NewWithConfig(context.Background(), config)
			if err != nil {
				t.Fatalf("create database pool: %v", err)
			}
			t.Cleanup(pool.Close)

			requestContext := context.WithValue(context.Background(), requestContextKey{}, requestContextMarker)
			requestContext, cancel := context.WithCancel(requestContext)
			t.Cleanup(cancel)

			request := httptest.NewRequest(http.MethodGet, tt.Path, nil).WithContext(requestContext)
			request = mux.SetURLVars(request, tt.RouteVars)
			response := httptest.NewRecorder()
			handlerDone := make(chan any, 1)
			go func() {
				var panicValue any
				defer func() {
					panicValue = recover()
					handlerDone <- panicValue
				}()
				tt.Handler(pool).ServeHTTP(response, request)
			}()

			select {
			case databaseContext := <-observedContext:
				if marker := databaseContext.Value(requestContextKey{}); marker != requestContextMarker {
					t.Errorf("database context marker = %v, want %q", marker, requestContextMarker)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("handler did not start a database query")
			}

			cancel()
			select {
			case panicValue := <-handlerDone:
				if panicValue != nil {
					t.Fatalf("handler panicked after cancellation: %v", panicValue)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("handler did not stop after request cancellation")
			}

			if response.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
			}
			if body := strings.TrimSpace(response.Body.String()); body != http.StatusText(http.StatusInternalServerError) {
				t.Fatalf("body = %q, want %q", body, http.StatusText(http.StatusInternalServerError))
			}
		})
	}
}
