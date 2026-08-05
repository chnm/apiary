package naturalearth

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlersPropagateRequestCancellation(t *testing.T) {
	testsupport.TestRequestCancellation(t, []testsupport.CancellationCase{
		{Name: "globe", Path: "/ne/globe", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return New(pool).NaturalEarthHandler() }},
		{Name: "filtered globe", Path: "/ne/globe?location=Europe&location=Asia", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return New(pool).NaturalEarthHandler() }},
	})
}
