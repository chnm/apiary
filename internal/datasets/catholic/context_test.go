package catholic

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlersPropagateRequestCancellation(t *testing.T) {
	testsupport.TestRequestCancellation(t, []testsupport.CancellationCase{
		{Name: "dioceses", Path: "/catholic-dioceses/", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return New(pool).CatholicDiocesesHandler() }},
		{Name: "dioceses per decade", Path: "/catholic-dioceses/per-decade/", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return New(pool).CatholicDiocesesPerDecadeHandler() }},
	})
}
