package apiary

import (
	"net/http"
	"os"
	"time"

	"github.com/chnm/apiary/internal/httpx"
	paramx "github.com/chnm/apiary/internal/params"
)

// internalServerError logs an internal error and returns a generic response to
// the client so database and encoding details are never exposed.
func internalServerError(w http.ResponseWriter, operation string, err error) {
	httpx.InternalServerError(w, operation, err)
}

// writeJSONResponse marshals a response before writing headers, allowing
// encoding failures to return a generic 500 response. Write failures can only
// be logged because the response may already be partially sent.
func writeJSONResponse(w http.ResponseWriter, value any) {
	httpx.WriteJSON(w, value)
}

// getEnv either returns the value of an environment variable or, if that
// environment variables does not exist, returns the fallback value provided.
func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

// dateInRange takes in a string which should be parsed to a date. That date is
// then kept within the range of the min and max dates passed as arguments.
func dateInRange(d string, min, max time.Time) (time.Time, error) {
	return paramx.DateInRange(d, min, max)
}

// intsToString formats a slice of ints as a comma-separated list, e.g. "152, 999".
func intsToString(ids []int) string {
	return httpx.IntsToString(ids)
}

type NullInt64 = httpx.NullInt64
type NullString = httpx.NullString
