// Package apb implements America's Public Bible endpoints.
package apb

import (
	"net/http"

	"github.com/chnm/apiary/internal/httpx"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/apb/bible-books", h.APBBibleBooksHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/bible-similarity", h.APBBibleSimilarityHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/bible-trend", h.APBBibleTrendHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/index/featured", h.APBIndexFeaturedHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/index/top", h.APBIndexTopHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/index/biblical", h.APBIndexBiblicalOrderHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/index/peaks", h.APBIndexChronologicalHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/index/all", h.APBIndexAllHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/verse", h.APBVerseHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/verse-quotations", h.APBVerseQuotationsHandler()).Methods("GET", "HEAD")
	router.HandleFunc("/apb/verse-trend", h.APBVerseTrendHandler()).Methods("GET", "HEAD")
}

func internalServerError(w http.ResponseWriter, operation string, err error) {
	httpx.InternalServerError(w, operation, err)
}

func writeJSONResponse(w http.ResponseWriter, value any) { httpx.WriteJSON(w, value) }
