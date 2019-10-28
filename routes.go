package relecapi

import (
	"github.com/gorilla/handlers"
)

// Routes registers the handlers for the URLs that should be served.
func (s *Server) Routes() {
	s.Router.HandleFunc("/presbyterians/", s.PresbyteriansHandler())
	s.Router.HandleFunc("/", s.EndpointHandler())
}

// Middleware registers the middleware functions that should be used.
func (s *Server) Middleware() {
	s.Router.Use(corsMiddleware)
	s.Router.Use(handlers.CompressHandler) // gzip requests
	s.Router.Use(loggingMiddleware)
}
