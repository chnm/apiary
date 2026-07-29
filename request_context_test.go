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

type serviceRequestContextKey struct{}

const serviceRequestContextMarker = "service-request-context"

func TestServiceHandlersPropagateRequestCancellation(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		routeVars map[string]string
		handler   func(*Server) http.HandlerFunc
	}{
		{
			name: "Presbyterians",
			path: "/presbyterians/",
			handler: func(server *Server) http.HandlerFunc {
				return server.PresbyteriansHandler()
			},
		},
		{
			name: "Religious Census denomination families",
			path: "/relcensus/denomination-families",
			handler: func(server *Server) http.HandlerFunc {
				return server.RelCensusDenominationFamiliesHandler()
			},
		},
		{
			name: "Religious Census denominations",
			path: "/relcensus/denominations?family_relec=Baptist",
			handler: func(server *Server) http.HandlerFunc {
				return server.RelCensusDenominationsHandler()
			},
		},
		{
			name: "Religious Census city membership by denomination",
			path: "/relcensus/city-membership?year=1926&denomination=Church+of+God+in+Christ",
			handler: func(server *Server) http.HandlerFunc {
				return server.RelCensusCityMembershipHandler()
			},
		},
		{
			name: "Religious Census city membership by family",
			path: "/relcensus/city-membership?year=1926&denominationFamily=Pentecostal",
			handler: func(server *Server) http.HandlerFunc {
				return server.RelCensusCityMembershipHandler()
			},
		},
		{
			name: "Religious Census city membership aggregate",
			path: "/relcensus/city-membership?year=1926",
			handler: func(server *Server) http.HandlerFunc {
				return server.RelCensusCityMembershipHandler()
			},
		},
		{
			name: "Religious Census locations",
			path: "/relcensus/cities",
			handler: func(server *Server) http.HandlerFunc {
				return server.RelCensusLocationsHandler()
			},
		},
		{
			name: "Natural Earth globe",
			path: "/ne/globe",
			handler: func(server *Server) http.HandlerFunc {
				return server.NaturalEarthHandler()
			},
		},
		{
			name: "Natural Earth filtered globe",
			path: "/ne/globe?location=Europe&location=Asia",
			handler: func(server *Server) http.HandlerFunc {
				return server.NaturalEarthHandler()
			},
		},
		{
			name:      "populated-place counties in state",
			path:      "/pop-places/state/nc/county/",
			routeVars: map[string]string{"state": "nc"},
			handler: func(server *Server) http.HandlerFunc {
				return server.CountiesInState()
			},
		},
		{
			name:      "populated places in county",
			path:      "/pop-places/county/mas_middlesex/place/",
			routeVars: map[string]string{"county": "mas_middlesex"},
			handler: func(server *Server) http.HandlerFunc {
				return server.PlacesInCounty()
			},
		},
		{
			name:      "populated-place details",
			path:      "/pop-places/place/611119/",
			routeVars: map[string]string{"place": "611119"},
			handler: func(server *Server) http.HandlerFunc {
				return server.Place()
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
				case <-time.After(500 * time.Millisecond):
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
				serviceRequestContextKey{},
				serviceRequestContextMarker,
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
				if marker := databaseContext.Value(serviceRequestContextKey{}); marker != serviceRequestContextMarker {
					t.Errorf(
						"database context marker = %v, want %q",
						marker,
						serviceRequestContextMarker,
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
