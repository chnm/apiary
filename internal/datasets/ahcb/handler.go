// Package ahcb implements Atlas of Historical County Boundaries endpoints.
package ahcb

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/", h.AHCBCountiesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/id/{id:[a-z_,]+}/", h.AHCBCountiesByIDHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/state-code/{state-code:[a-z,]+}/", h.AHCBCountiesByStateCodeHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/state-terr-id/{state-terr-id:[a-z_,]+}/", h.AHCBCountiesByStateTerrIDHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/ahcb/states/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/", h.AHCBStatesHandler()).Methods("GET", "HEAD")
}
