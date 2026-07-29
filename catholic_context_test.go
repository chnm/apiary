package apiary

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type catholicRequestContextKey struct{}

const catholicRequestContextMarker = "catholic-request-context"

func TestCatholicHandlersPropagateRequestCancellation(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		handler func(*Server) http.HandlerFunc
	}{
		{
			name: "dioceses",
			path: "/catholic-dioceses/",
			handler: func(server *Server) http.HandlerFunc {
				return server.CatholicDiocesesHandler()
			},
		},
		{
			name: "dioceses per decade",
			path: "/catholic-dioceses/per-decade/",
			handler: func(server *Server) http.HandlerFunc {
				return server.CatholicDiocesesPerDecadeHandler()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observedContext := make(chan context.Context, 1)
			config, err := pgxpool.ParseConfig(
				"postgres://apiary:apiary@127.0.0.1:1/apiary",
			)
			if err != nil {
				t.Fatalf("parse database config: %v", err)
			}
			config.BeforeConnect = func(ctx context.Context, _ *pgx.ConnConfig) error {
				observedContext <- ctx
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(2 * time.Second):
					return errors.New("database context was not canceled")
				}
			}

			pool, err := pgxpool.NewWithConfig(context.Background(), config)
			if err != nil {
				t.Fatalf("create database pool: %v", err)
			}
			t.Cleanup(pool.Close)

			requestContext := context.WithValue(
				context.Background(),
				catholicRequestContextKey{},
				catholicRequestContextMarker,
			)
			requestContext, cancel := context.WithCancel(requestContext)
			t.Cleanup(cancel)

			request := httptest.NewRequest(http.MethodGet, tt.path, nil).
				WithContext(requestContext)
			response := httptest.NewRecorder()
			handlerDone := make(chan any, 1)

			go func() {
				var panicValue any
				defer func() {
					panicValue = recover()
					handlerDone <- panicValue
				}()
				tt.handler(&Server{DB: pool}).ServeHTTP(response, request)
			}()

			select {
			case databaseContext := <-observedContext:
				if marker := databaseContext.Value(catholicRequestContextKey{}); marker != catholicRequestContextMarker {
					t.Errorf(
						"database context marker = %v, want %q",
						marker,
						catholicRequestContextMarker,
					)
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
				t.Fatalf(
					"status = %d, want %d",
					response.Code,
					http.StatusInternalServerError,
				)
			}
			if body := strings.TrimSpace(response.Body.String()); body != http.StatusText(http.StatusInternalServerError) {
				t.Fatalf(
					"body = %q, want %q",
					body,
					http.StatusText(http.StatusInternalServerError),
				)
			}
		})
	}
}
