package apiary

import (
	"github.com/chnm/apiary/internal/datasets/ahcb"
	"github.com/chnm/apiary/internal/datasets/apb"
	"github.com/chnm/apiary/internal/datasets/bom"
	"github.com/chnm/apiary/internal/datasets/catholic"
	"github.com/chnm/apiary/internal/datasets/naturalearth"
	"github.com/chnm/apiary/internal/datasets/pinkertons"
	"github.com/chnm/apiary/internal/datasets/popplaces"
	"github.com/chnm/apiary/internal/datasets/presbyterians"
	"github.com/chnm/apiary/internal/datasets/relcensus"
)

// Routes registers the handlers for the URLs that should be served.
func (s *Server) Routes() {
	ahcb.New(s.DB).RegisterRoutes(s.Router)
	apb.New(s.DB).RegisterRoutes(s.Router)
	bom.New(s.DB).RegisterRoutes(s.Router)
	catholic.New(s.DB).RegisterRoutes(s.Router)
	naturalearth.New(s.DB).RegisterRoutes(s.Router)
	popplaces.New(s.DB).RegisterRoutes(s.Router)
	presbyterians.New(s.DB).RegisterRoutes(s.Router)
	relcensus.New(s.DB).RegisterRoutes(s.Router)
	pinkertons.New(s.DB).RegisterRoutes(s.Router)
	s.Router.HandleFunc("/", s.EndpointsHandler()).Methods("GET", "HEAD")

	// Make sure to log 404 errors
	if s.Config.logging {
		s.Router.NotFoundHandler = loggingMiddleware(s.NotFoundHandler())
	} else {
		s.Router.NotFoundHandler = s.NotFoundHandler()
	}
}
