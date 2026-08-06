// Package relcensus implements Religious Bodies Census endpoints.
package relcensus

import (
	"github.com/chnm/apiary/internal/httpx"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/relcensus/denomination-families", h.RelCensusDenominationFamiliesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/relcensus/denominations", h.RelCensusDenominationsHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/relcensus/city-membership", h.RelCensusCityMembershipHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/relcensus/cities", h.RelCensusLocationsHandler()).Methods("GET", "HEAD")
}

type NullInt64 = httpx.NullInt64
type NullString = httpx.NullString
