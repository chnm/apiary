package apiary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CacheTest returns the time the application started and the time that the
// handler was last run. So if the result is cached, then one would expect that
// the time the handler was last run would remain the same.
func (s *Server) CacheTest() http.HandlerFunc {
	startup := time.Now() // This will be captured by the closure at application startup time
	return func(w http.ResponseWriter, r *http.Request) {
		out := struct {
			Startup time.Time `json:"startup"`
			Handler time.Time `json:"handler"`
		}{
			Startup: startup,
			Handler: time.Now(), // This will be the time the handler was run
		}
		response, _ := json.Marshal(out)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}
