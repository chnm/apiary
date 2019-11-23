package relecapi

// Routes registers the handlers for the URLs that should be served.
func (s *Server) Routes() {
	s.Router.HandleFunc("/presbyterians/", s.PresbyteriansHandler()).Methods("GET")
	s.Router.HandleFunc("/ahcb/states/{date}/", s.AHCBStatesHandler()).Methods("GET")
	s.Router.HandleFunc("/", s.EndpointHandler()).Methods("GET")

	// Make sure to log 404 errors
	if getEnv("RELECAPI_LOGGING", "on") == "on" {
		s.Router.NotFoundHandler = loggingMiddleware(s.NotFoundHandler())
	} else {
		s.Router.NotFoundHandler = s.NotFoundHandler()
	}
}
