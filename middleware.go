package relecapi

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

// Middleware registers the middleware functions that should be used.
func (s *Server) Middleware() {
	if getEnv("RELECAPI_LOGGING", "on") == "on" {
		s.Router.Use(loggingMiddleware)
	}
	s.Router.Use(corsMiddleware)
	s.Router.Use(handlers.CompressHandler)   // gzip requests
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

// NotFoundHandler returns 404 errors
func (s *Server) NotFoundHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not found."))
	})

}
