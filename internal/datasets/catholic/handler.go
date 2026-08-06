// Package catholic implements the Catholic dioceses endpoints.
package catholic

import (
	"github.com/chnm/apiary/internal/httpx"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/catholic-dioceses/", h.CatholicDiocesesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/catholic-dioceses/per-decade/", h.CatholicDiocesesPerDecadeHandler()).Methods("GET", "HEAD")
}

type NullInt64 = httpx.NullInt64
