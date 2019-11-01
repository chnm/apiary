package relecapi

// Routes registers the handlers for the URLs that should be served.
func (s *Server) Routes() {
	s.Router.HandleFunc("/presbyterians/", s.PresbyteriansHandler())
	s.Router.HandleFunc("/", s.EndpointHandler())
	s.Router.NotFoundHandler = loggingMiddleware(s.NotFoundHandler())
}
