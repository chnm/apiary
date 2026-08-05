// Package pinkertons implements Pinkerton surveillance-data endpoints.
package pinkertons

import (
	"github.com/chnm/apiary/internal/httpx"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/pinkertons/activities", h.ActivitiesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/pinkertons/activities/{id:[0-9]+}", h.ActivityByIDHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/pinkertons/locations", h.LocationsHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/pinkertons/operatives", h.OperativesHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/pinkertons/subjects", h.SubjectsHandler()).Methods("GET", "HEAD")
}

type NullInt64 = httpx.NullInt64
type NullString = httpx.NullString
