package pinkertons

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

type pinkertonsRequestContextKey struct{}

const pinkertonsRequestContextMarker = "pinkertons-request-context"

func newPinkertonsCancellationPool(
	t *testing.T,
	observedContext chan<- context.Context,
) *pgxpool.Pool {
	t.Helper()

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
	return pool
}

func newPinkertonsRequest(t *testing.T, path string) (*http.Request, context.CancelFunc) {
	t.Helper()

	requestContext := context.WithValue(
		context.Background(),
		pinkertonsRequestContextKey{},
		pinkertonsRequestContextMarker,
	)
	requestContext, cancel := context.WithCancel(requestContext)
	request := httptest.NewRequest(http.MethodGet, path, nil).
		WithContext(requestContext)
	return request, cancel
}

func assertPinkertonsDatabaseContext(
	t *testing.T,
	observedContext <-chan context.Context,
) {
	t.Helper()

	select {
	case databaseContext := <-observedContext:
		if marker := databaseContext.Value(pinkertonsRequestContextKey{}); marker != pinkertonsRequestContextMarker {
			t.Errorf("database context marker = %v, want %q", marker, pinkertonsRequestContextMarker)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("database query did not start")
	}
}

func TestPinkertonsHandlersPropagateRequestCancellation(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		routeVars map[string]string
		handler   func(*Handler) http.HandlerFunc
	}{
		{
			name: "activities",
			path: "/pinkertons/activities?limit=1",
			handler: func(server *Handler) http.HandlerFunc {
				return server.ActivitiesHandler()
			},
		},
		{
			name:      "activity by ID",
			path:      "/pinkertons/activities/1",
			routeVars: map[string]string{"id": "1"},
			handler: func(server *Handler) http.HandlerFunc {
				return server.ActivityByIDHandler()
			},
		},
		{
			name: "locations",
			path: "/pinkertons/locations",
			handler: func(server *Handler) http.HandlerFunc {
				return server.LocationsHandler()
			},
		},
		{
			name: "operatives",
			path: "/pinkertons/operatives",
			handler: func(server *Handler) http.HandlerFunc {
				return server.OperativesHandler()
			},
		},
		{
			name: "subjects",
			path: "/pinkertons/subjects",
			handler: func(server *Handler) http.HandlerFunc {
				return server.SubjectsHandler()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observedContext := make(chan context.Context, 1)
			pool := newPinkertonsCancellationPool(t, observedContext)
			request, cancel := newPinkertonsRequest(t, tt.path)
			t.Cleanup(cancel)
			if tt.routeVars != nil {
				request = mux.SetURLVars(request, tt.routeVars)
			}

			response := httptest.NewRecorder()
			handlerDone := make(chan any, 1)
			go func() {
				var panicValue any
				defer func() {
					panicValue = recover()
					handlerDone <- panicValue
				}()
				tt.handler(&Handler{db: pool}).ServeHTTP(response, request)
			}()

			assertPinkertonsDatabaseContext(t, observedContext)
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

func TestActivityLocationsPropagatesRequestCancellation(t *testing.T) {
	observedContext := make(chan context.Context, 1)
	pool := newPinkertonsCancellationPool(t, observedContext)
	request, cancel := newPinkertonsRequest(t, "/pinkertons/activities?limit=1")
	t.Cleanup(cancel)

	queryDone := make(chan error, 1)
	go func() {
		_, err := (&Handler{db: pool}).activityLocations(request.Context(), 1)
		queryDone <- err
	}()

	assertPinkertonsDatabaseContext(t, observedContext)
	cancel()

	select {
	case err := <-queryDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("activityLocations error = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("location query did not stop after request cancellation")
	}
}

func TestBulkActivityLocationsPropagatesRequestCancellation(t *testing.T) {
	observedContext := make(chan context.Context, 1)
	pool := newPinkertonsCancellationPool(t, observedContext)
	request, cancel := newPinkertonsRequest(t, "/pinkertons/activities?limit=2")
	t.Cleanup(cancel)

	queryDone := make(chan error, 1)
	go func() {
		_, err := (&Handler{db: pool}).activityLocationsByActivityIDs(
			request.Context(),
			[]int{1, 2},
		)
		queryDone <- err
	}()

	assertPinkertonsDatabaseContext(t, observedContext)
	cancel()

	select {
	case err := <-queryDone:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf(
				"activityLocationsByActivityIDs error = %v, want context.Canceled",
				err,
			)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("bulk location query did not stop after request cancellation")
	}
}
