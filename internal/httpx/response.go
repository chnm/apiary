package httpx

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// InternalServerError logs an internal error and returns a generic response.
func InternalServerError(w http.ResponseWriter, operation string, err error) {
	log.Printf("%s: %v", operation, err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// WriteJSON marshals value before writing headers so encoding errors can still
// produce a generic error response.
func WriteJSON(w http.ResponseWriter, value any) {
	response, err := json.Marshal(value)
	if err != nil {
		InternalServerError(w, "error marshaling JSON response", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(response); err != nil {
		log.Printf("error writing JSON response: %v", err)
	}
}

// IntsToString formats integers as a comma-separated list.
func IntsToString(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ", ")
}
