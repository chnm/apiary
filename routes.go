package relecapi

// Routes registers the handlers for the URLs that should be served.
func (s *Server) Routes() {
	s.Router.HandleFunc("/presbyterians/", s.PresbyteriansHandler())
}
