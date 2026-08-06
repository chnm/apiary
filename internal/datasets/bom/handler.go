// Package bom implements the Bills of Mortality dataset endpoints.
package bom

import (
	"net/http"

	"github.com/chnm/apiary/internal/httpx"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler owns the dependencies and HTTP handlers for the BOM dataset.
type Handler struct {
	db *pgxpool.Pool
}

// New creates a BOM dataset handler.
func New(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers all BOM routes.
func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/bom/parishes", h.ParishesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/bom/totalbills", h.TotalBillsHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/bom/statistics", h.StatisticsHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/bom/bills", h.BillsHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/bom/shapefiles", h.BillsShapefilesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/bom/christenings", h.ChristeningsHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/bom/causes", h.DeathCausesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/bom/list-deaths", h.ListCausesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/bom/list-christenings", h.ListChristeningsHandler()).Methods("GET", "HEAD")
}

type NullInt64 = httpx.NullInt64
type NullString = httpx.NullString

func internalServerError(w http.ResponseWriter, operation string, err error) {
	httpx.InternalServerError(w, operation, err)
}

func writeJSONResponse(w http.ResponseWriter, value any) {
	httpx.WriteJSON(w, value)
}

func intsToString(ids []int) string {
	return httpx.IntsToString(ids)
}
