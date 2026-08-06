package apb

import "net/http"

const singleReferenceError = "400 Bad request. Please provide exactly one reference."

func requireSingleReference(w http.ResponseWriter, r *http.Request) (string, bool) {
	references := r.URL.Query()["ref"]
	if len(references) != 1 {
		http.Error(w, singleReferenceError, http.StatusBadRequest)
		return "", false
	}

	return references[0], true
}
