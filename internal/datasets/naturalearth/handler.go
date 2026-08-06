// Package naturalearth implements Natural Earth endpoints.
package naturalearth

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ne/globe", h.NaturalEarthHandler()).Methods("GET", "HEAD")
}
