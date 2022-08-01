package apiary

// Routes registers the handlers for the URLs that should be served.
func (s *Server) Routes() {
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/", s.AHCBCountiesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/id/{id:[a-z_,]+}/", s.AHCBCountiesByIDHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/state-code/{state-code:[a-z,]+}/", s.AHCBCountiesByStateCodeHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/counties/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/state-terr-id/{state-terr-id:[a-z_,]+}/", s.AHCBCountiesByStateTerrIDHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ahcb/states/{date:[0-9]{4}-[0-9]{2}-[0-9]{2}}/", s.AHCBStatesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/bible-books", s.APBBibleBooksHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/bible-similarity", s.APBBibleSimilarityHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/bible-trend", s.APBBibleTrendHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/index/featured", s.APBIndexFeaturedHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/index/top", s.APBIndexTopHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/index/biblical", s.APBIndexBiblicalOrderHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/index/peaks", s.APBIndexChronologicalHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/index/all", s.APBIndexAllHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/verse", s.APBVerseHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/verse-quotations", s.APBVerseQuotationsHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/apb/verse-trend", s.APBVerseTrendHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/bom/parishes", s.ParishesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/bom/totalbills", s.TotalBillsHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/bom/bills", s.BillsHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/bom/christenings", s.ChristeningsHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/bom/causes", s.DeathCausesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/bom/list-deaths", s.ListCausesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/bom/list-christenings", s.ListChristeningsHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/catholic-dioceses/", s.CatholicDiocesesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/catholic-dioceses/per-decade/", s.CatholicDiocesesPerDecadeHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/ne/northamerica/", s.NENorthAmericaHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/pop-places/county/{county:[a-z_,]+}/place/", s.PlacesInCounty()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/pop-places/place/{place}/", s.Place()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/pop-places/state/{state:[a-z]{2}}/county/", s.CountiesInState()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/presbyterians/", s.PresbyteriansHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/relcensus/denomination-families", s.RelCensusDenominationFamiliesHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/relcensus/denominations", s.RelCensusDenominationsHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/relcensus/city-membership", s.RelCensusCityMembershipHandler()).Methods("GET", "HEAD")
	s.Router.HandleFunc("/", s.EndpointsHandler()).Methods("GET", "HEAD")

	// Make sure to log 404 errors
	if s.Config.logging {
		s.Router.NotFoundHandler = loggingMiddleware(s.NotFoundHandler())
	} else {
		s.Router.NotFoundHandler = s.NotFoundHandler()
	}
}
