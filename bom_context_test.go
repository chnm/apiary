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

type bomRequestContextKey struct{}

const bomRequestContextMarker = "bom-request-context"

func TestBOMHandlersPropagateRequestCancellation(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		handler func(*Server) http.HandlerFunc
	}{
		{
			name: "bills",
			path: "/bom/bills?start-year=1669&end-year=1670&bill-type=weekly",
			handler: func(server *Server) http.HandlerFunc {
				return server.BillsHandler()
			},
		},
		{
			name: "total bills",
			path: "/bom/totalbills?type=weekly",
			handler: func(server *Server) http.HandlerFunc {
				return server.TotalBillsHandler()
			},
		},
		{
			name: "statistics",
			path: "/bom/statistics?type=yearly",
			handler: func(server *Server) http.HandlerFunc {
				return server.StatisticsHandler()
			},
		},
		{
			name: "death causes",
			path: "/bom/causes?start-year=1669&end-year=1670",
			handler: func(server *Server) http.HandlerFunc {
				return server.DeathCausesHandler()
			},
		},
		{
			name: "list causes",
			path: "/bom/list-deaths",
			handler: func(server *Server) http.HandlerFunc {
				return server.ListCausesHandler()
			},
		},
		{
			name: "christenings",
			path: "/bom/christenings?start-year=1669&end-year=1670",
			handler: func(server *Server) http.HandlerFunc {
				return server.ChristeningsHandler()
			},
		},
		{
			name: "list christenings",
			path: "/bom/list-christenings",
			handler: func(server *Server) http.HandlerFunc {
				return server.ListChristeningsHandler()
			},
		},
		{
			name: "parishes",
			path: "/bom/parishes",
			handler: func(server *Server) http.HandlerFunc {
				return server.ParishesHandler()
			},
		},
		{
			name: "parish geometries",
			path: "/bom/geometries",
			handler: func(server *Server) http.HandlerFunc {
				return server.ParishShpHandler()
			},
		},
		{
			name: "bill geometries",
			path: "/bom/shapefiles?start-year=1669&end-year=1670&bill-type=weekly&count-type=plague",
			handler: func(server *Server) http.HandlerFunc {
				return server.BillsShapefilesHandler()
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
				bomRequestContextKey{},
				bomRequestContextMarker,
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
				if marker := databaseContext.Value(bomRequestContextKey{}); marker != bomRequestContextMarker {
					t.Errorf("database context marker = %v, want %q", marker, bomRequestContextMarker)
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
			if tt.name == "statistics" {
				if body := strings.TrimSpace(response.Body.String()); body != http.StatusText(http.StatusInternalServerError) {
					t.Fatalf(
						"body = %q, want %q",
						body,
						http.StatusText(http.StatusInternalServerError),
					)
				}
			}
		})
	}
}
