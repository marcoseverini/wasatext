package api

import (
	"encoding/json"
	"net/http"
)

func (rt *_router) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	errRes := ErrorResponse{
		Message: message,
	}
	_ = json.NewEncoder(w).Encode(errRes)
}

