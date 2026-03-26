package api

import (
	"encoding/json"
	"gitlab.pg.innopolis.university/e.pustovoytenko/dnp25-project-19/pkg/logger"
	"go.uber.org/zap"
	"net/http"
)

// ErrorResponse is util function that sends json with information about error and the corresponding status code
func ErrorResponse(w http.ResponseWriter, status int, msg string) {
	logger.Info("Client's request failed",
		zap.Int("HTTP code", status),
		zap.String("Error", msg))
	ResponseJson(w, status, map[string]string{"error": msg})
}

// ResponseJson is util function that forms the json-formatted response, sets the corresponding headers
// and status code. [payload] is the object to be encoded in json format.
func ResponseJson(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		payloadStr, _ := json.Marshal(payload)

	  logger.Info("Successful response",
		zap.String("Payload", string(payloadStr)))
	}
}
