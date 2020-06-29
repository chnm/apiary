package dataapi

// Routes registers the handlers for the URLs that should be served.
func (s *Server) Routes() {
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/", s.AHCBCountiesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/id/{id:[a-z_,]+}/", s.AHCBCountiesByIDHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/state-terr-id/{state-terr-id:[a-z_,]+}/", s.AHCBCountiesByStateTerrIDHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/state-code/{state-code:[a-z,]+}/", s.AHCBCountiesByStateCodeHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/states/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/", s.AHCBStatesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/catholic-dioceses/", s.CatholicDiocesesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/catholic-dioceses/per-decade/", s.CatholicDiocesesPerDecadeHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ne/northamerica/", s.NENorthAmericaHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/pop-places/state/{state:[a-z]{2}}/county/", s.CountiesInState()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/pop-places/county/{county:[a-z_,]+}/place/", s.PlacesInCounty()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/pop-places/place/{place}/", s.Place()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/presbyterians/", s.PresbyteriansHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/verse", s.VerseHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/verse-trend", s.VerseTrendHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/", s.EndpointsHandler()).Methods("GET", "HEAD")

	// Make sure to log 404 errors
	if s.Config.logging {
		s.Router.NotFoundHandler = loggingMiddleware(s.NotFoundHandler())
	} else {
		s.Router.NotFoundHandler = s.NotFoundHandler()
	}
}
