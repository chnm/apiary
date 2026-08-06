package presbyterians

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlersPropagateRequestCancellation(t *testing.T) {
	testsupport.TestRequestCancellation(t, []testsupport.CancellationCase{{
		Name: "statistics", Path: "/presbyterians/", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return New(pool).PresbyteriansHandler() },
	}})
}
