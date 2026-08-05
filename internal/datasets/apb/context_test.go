package apb

import (
	"net/http"
	"testing"

	"github.com/chnm/apiary/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestHandlersPropagateRequestCancellation(t *testing.T) {
	newHandler := func(pool *pgxpool.Pool) *Handler { return New(pool) }
	testsupport.TestRequestCancellation(t, []testsupport.CancellationCase{
		{Name: "featured index", Path: "/apb/index/featured", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBIndexFeaturedHandler() }},
		{Name: "biblical index", Path: "/apb/index/biblical", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBIndexBiblicalOrderHandler() }},
		{Name: "top index", Path: "/apb/index/top", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBIndexTopHandler() }},
		{Name: "chronological index", Path: "/apb/index/peaks", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBIndexChronologicalHandler() }},
		{Name: "all index", Path: "/apb/index/all", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBIndexAllHandler() }},
		{Name: "bible books", Path: "/apb/bible-books", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBBibleBooksHandler() }},
		{Name: "bible similarity", Path: "/apb/bible-similarity", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBBibleSimilarityHandler() }},
		{Name: "bible trend", Path: "/apb/bible-trend", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBBibleTrendHandler() }},
		{Name: "verse", Path: "/apb/verse?ref=Genesis+1:1", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBVerseHandler() }},
		{Name: "verse trend", Path: "/apb/verse-trend?ref=Genesis+1:1&corpus=chronam", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBVerseTrendHandler() }},
		{Name: "verse quotations", Path: "/apb/verse-quotations?ref=Genesis+1:1", Handler: func(pool *pgxpool.Pool) http.HandlerFunc { return newHandler(pool).APBVerseQuotationsHandler() }},
	})
}
