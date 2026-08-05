package popplaces

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlersPropagateRequestCancellation(t *testing.T) {
	testsupport.TestRequestCancellation(t, []testsupport.CancellationCase{
		{Name: "counties in state", Path: "/pop-places/state/nc/county/", RouteVars: map[string]string{"state": "nc"}, Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return New(pool).CountiesInState() }},
		{Name: "places in county", Path: "/pop-places/county/mas_middlesex/place/", RouteVars: map[string]string{"county": "mas_middlesex"}, Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return New(pool).PlacesInCounty() }},
		{Name: "place details", Path: "/pop-places/place/611119/", RouteVars: map[string]string{"place": "611119"}, Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return New(pool).Place() }},
	})
}
