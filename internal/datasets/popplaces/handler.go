// Package popplaces implements populated-place endpoints.
package popplaces

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/pop-places/county/{county:[a-z_,]+}/place/", h.PlacesInCounty()).Methods("GET", "HEAD")
	router.HandleFunc("/pop-places/place/{place}/", h.Place()).Methods("GET", "HEAD")
	router.HandleFunc("/pop-places/state/{state:[a-z]{2}}/county/", h.CountiesInState()).Methods("GET", "HEAD")
}
