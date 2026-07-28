package apiary

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

type ahcbRequestContextKey struct{}

const ahcbRequestContextMarker = "ahcb-request-context"

func TestAHCBHandlersPropagateRequestCancellation(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		routeVars map[string]string
		handler   func(*Server) http.HandlerFunc
	}{
		{
			name:      "states",
			path:      "/ahcb/states/1789-07-04/",
			routeVars: map[string]string{"date": "1789-07-04"},
			handler: func(server *Server) http.HandlerFunc {
				return server.AHCBStatesHandler()
			},
		},
		{
			name:      "counties",
			path:      "/ahcb/counties/1926-07-04/",
			routeVars: map[string]string{"date": "1926-07-04"},
			handler: func(server *Server) http.HandlerFunc {
				return server.AHCBCountiesHandler()
			},
		},
		{
			name: "counties by ID",
			path: "/ahcb/counties/1980-12-31/id/vas_fairfax,vas_arlington/",
			routeVars: map[string]string{
				"date": "1980-12-31",
				"id":   "vas_fairfax,vas_arlington",
			},
			handler: func(server *Server) http.HandlerFunc {
				return server.AHCBCountiesByIDHandler()
			},
		},
		{
			name: "counties by state territory ID",
			path: "/ahcb/counties/1980-12-31/state-terr-id/ga_state,va_state/",
			routeVars: map[string]string{
				"date":          "1980-12-31",
				"state-terr-id": "ga_state,va_state",
			},
			handler: func(server *Server) http.HandlerFunc {
				return server.AHCBCountiesByStateTerrIDHandler()
			},
		},
		{
			name: "counties by state code",
			path: "/ahcb/counties/1940-12-31/state-code/nd,sd/",
			routeVars: map[string]string{
				"date":       "1940-12-31",
				"state-code": "nd,sd",
			},
			handler: func(server *Server) http.HandlerFunc {
				return server.AHCBCountiesByStateCodeHandler()
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
				ahcbRequestContextKey{},
				ahcbRequestContextMarker,
			)
			requestContext, cancel := context.WithCancel(requestContext)
			t.Cleanup(cancel)

			request := httptest.NewRequest(http.MethodGet, tt.path, nil).
				WithContext(requestContext)
			request = mux.SetURLVars(request, tt.routeVars)
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
				if marker := databaseContext.Value(ahcbRequestContextKey{}); marker != ahcbRequestContextMarker {
					t.Errorf(
						"database context marker = %v, want %q",
						marker,
						ahcbRequestContextMarker,
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
