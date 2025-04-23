package api

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is util function that sends json with information about error and the corresponding status code
func ErrorResponse(w http.ResponseWriter, status int, msg string) {
	ResponseJson(w, status, map[string]string{"error": msg})
}

// ResponseJson is util function that forms the json-formatted response, sets the corresponding headers
// and status code. [payload] is the object to be encoded in json format.
func ResponseJson(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
