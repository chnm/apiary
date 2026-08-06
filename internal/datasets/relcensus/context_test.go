package relcensus

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlersPropagateRequestCancellation(t *testing.T) {
	newHandler := func(pool *pgxpool.Pool) *Handler { return New(pool) }
	testsupport.TestRequestCancellation(t, []testsupport.CancellationCase{
		{Name: "denomination families", Path: "/relcensus/denomination-families", Handler: func(pool *pgxpool.Pool) http.HandlerFunc {
			return newHandler(pool).RelCensusDenominationFamiliesHandler()
		}},
		{Name: "denominations", Path: "/relcensus/denominations?family_relec=Baptist", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).RelCensusDenominationsHandler() }},
		{Name: "membership by denomination", Path: "/relcensus/city-membership?year=1926&denomination=Church+of+God+in+Christ", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).RelCensusCityMembershipHandler() }},
		{Name: "membership by family", Path: "/relcensus/city-membership?year=1926&denominationFamily=Pentecostal", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).RelCensusCityMembershipHandler() }},
		{Name: "aggregate membership", Path: "/relcensus/city-membership?year=1926", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).RelCensusCityMembershipHandler() }},
		{Name: "locations", Path: "/relcensus/cities", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).RelCensusLocationsHandler() }},
	})
}
