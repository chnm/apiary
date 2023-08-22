package apiary

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

// Middleware registers the middleware functions that should be used.
func (s *Server) Middleware() {
	if s.Config.logging {
		s.Router.Use(loggingMiddleware)
	}
	s.Router.Use(corsMiddleware)
	s.Router.Use(clientCacheMiddleware)
	s.Router.Use(handlers.CompressHandler) // gzip requests
	s.Router.Use(s.Cache.Middleware)
	s.Router.Use(handlers.RecoveryHandler()) // Recover from runtime panics
}

// Log requests in the Apache Common Log format
func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}

// Allow Cross-Origin Request Sharing
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

// clientCacheMiddleware sets HTTP headers to permit client-side caching.
func clientCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=604800") // One week
		next.ServeHTTP(w, r)
	})
}

// NotFoundHandler returns 404 errors
func (s *Server) NotFoundHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not found."))
	})

}
