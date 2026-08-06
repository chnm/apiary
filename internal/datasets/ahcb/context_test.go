package ahcb

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlersPropagateRequestCancellation(t *testing.T) {
	newHandler := func(pool *pgxpool.Pool) *Handler { return New(pool) }
	testsupport.TestRequestCancellation(t, []testsupport.CancellationCase{
		{Name: "states", Path: "/ahcb/states/1789-07-04/", RouteVars: map[string]string{"date": "1789-07-04"}, Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).AHCBStatesHandler() }},
		{Name: "counties", Path: "/ahcb/counties/1926-07-04/", RouteVars: map[string]string{"date": "1926-07-04"}, Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).AHCBCountiesHandler() }},
		{Name: "counties by ID", Path: "/ahcb/counties/1980-12-31/id/vas_fairfax,vas_arlington/", RouteVars: map[string]string{"date": "1980-12-31", "id": "vas_fairfax,vas_arlington"}, Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).AHCBCountiesByIDHandler() }},
		{Name: "counties by state territory ID", Path: "/ahcb/counties/1980-12-31/state-terr-id/ga_state,va_state/", RouteVars: map[string]string{"date": "1980-12-31", "state-terr-id": "ga_state,va_state"}, Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).AHCBCountiesByStateTerrIDHandler() }},
		{Name: "counties by state code", Path: "/ahcb/counties/1940-12-31/state-code/nd,sd/", RouteVars: map[string]string{"date": "1940-12-31", "state-code": "nd,sd"}, Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).AHCBCountiesByStateCodeHandler() }},
	})
}
