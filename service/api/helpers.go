package api

import (
	"encoding/json" // Libreria per codificare/decodificare JSON
	"net/http" // Strumenti per gestire l'HTTP
)

// Funzione per inviare una risposta di errore JSON con il codice di stato e il messaggio specificati
func (rt *_router) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	errRes := ErrorResponse{
		Message: message,
	}
	_ = json.NewEncoder(w).Encode(errRes)
}

