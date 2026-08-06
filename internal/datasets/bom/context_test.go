package bom

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlersPropagateRequestCancellation(t *testing.T) {
	newHandler := func(pool *pgxpool.Pool) *Handler { return New(pool) }
	testsupport.TestRequestCancellation(t, []testsupport.CancellationCase{
		{Name: "bills", Path: "/bom/bills?start-year=1669&end-year=1670&bill-type=weekly", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).BillsHandler() }},
		{Name: "total bills", Path: "/bom/totalbills?type=weekly", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).TotalBillsHandler() }},
		{Name: "statistics", Path: "/bom/statistics?type=yearly", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).StatisticsHandler() }},
		{Name: "death causes", Path: "/bom/causes?start-year=1669&end-year=1670", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).DeathCausesHandler() }},
		{Name: "list causes", Path: "/bom/list-deaths", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).ListCausesHandler() }},
		{Name: "christenings", Path: "/bom/christenings?start-year=1669&end-year=1670", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).ChristeningsHandler() }},
		{Name: "list christenings", Path: "/bom/list-christenings", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).ListChristeningsHandler() }},
		{Name: "parishes", Path: "/bom/parishes", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).ParishesHandler() }},
		{Name: "bill geometries", Path: "/bom/shapefiles?start-year=1669&end-year=1670&bill-type=weekly&count-type=plague", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).BillsShapefilesHandler() }},
	})
}
