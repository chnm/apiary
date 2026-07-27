package apiary

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDeathCausesHandlerQueryErrorReturnsInternalServerError(t *testing.T) {
	pool, err := pgxpool.New(
		context.Background(),
		"postgres://apiary:apiary@127.0.0.1:1/apiary",
	)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	pool.Close()

	server := &Server{DB: pool}
	request := httptest.NewRequest(
		http.MethodGet,
		"/bom/causes?start-year=1669&end-year=1754",
		nil,
	)
	response := httptest.NewRecorder()

	server.DeathCausesHandler().ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}

	if body := strings.TrimSpace(response.Body.String()); body != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("body = %q, want %q", body, http.StatusText(http.StatusInternalServerError))
	}
}
