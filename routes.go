package dataapi

// Routes registers the handlers for the URLs that should be served.
func (s *Server) Routes() {
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/", s.AHCBCountiesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/id/{id:[a-z_,]+}/", s.AHCBCountiesByIdHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/state-terr-id/{state-terr-id:[a-z_,]+}/", s.AHCBCountiesByStateTerrIdHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/state-code/{state-code:[a-z,]+}/", s.AHCBCountiesByStateCodeHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/states/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/", s.AHCBStatesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/catholic-dioceses/", s.CatholicDiocesesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ne/northamerica/", s.NENorthAmericaHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/presbyterians/", s.PresbyteriansHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/", s.SourcesHandler()).Methods("GET")

	// Make sure to log 404 errors
	if s.Config.logging {
		s.Router.NotFoundHandler = loggingMiddleware(s.NotFoundHandler())
	} else {
		s.Router.NotFoundHandler = s.NotFoundHandler()
	}
}
