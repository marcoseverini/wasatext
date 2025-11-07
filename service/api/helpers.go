package api

import (
	"encoding/json" // Libreria per la codifica/decodifica JSON
	"net/http"      // Libreria per gestire HTTP
)

// Invia una risposta di errore JSON con il codice di stato e il messaggio specificati
func (rt *_router) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errRes := ErrorResponse{
		Message: message,
	}

	_ = json.NewEncoder(w).Encode(errRes)
}
