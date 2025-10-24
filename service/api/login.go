package api

import (
	"encoding/json" // traduttore JSON
	"net/http"      // Strumenti per gestire l'HTTP

	"github.com/julienschmidt/httprouter" // router HTTP di terze parti
)

// Handler per l'endpoint POST /session
// r è la richiesta JSON in entrata
// w è la risposta JSON in uscita
// _ sono i parametri dell'URL (non usati in questo endpoint)
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	
	// Leggiamo la richiesta JSON e la trasformiamo in una struct LoginRequest
	var req LoginRequest 
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "JSON non valido: "+err.Error(), http.StatusBadRequest)
		return 
	}

	// Controlliamo che il nome rispetti le regole del nostro api.yaml (min: 3, max: 16)
	if len(req.Name) < 3 || len(req.Name) > 16 {
		http.Error(w, "Nome utente non valido (deve essere tra 3 e 16 caratteri)", http.StatusBadRequest)
		return 
	}

	// Passiamo il nome utente al database per fare il login
	// Riceviamo indietro l'utente (esistente o appena creato), oppure un errore
	user, err := rt.db.DoLogin(req.Name)
	if err != nil {
		http.Error(w, "Errore interno del server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Prepariamo la risposta JSON (identificatore dell'utente) da inviare al client
	res := LoginResponse{
		Identifier: user.ID, 
	}

	// Impostiamo il codice di stato a "201 Created" 
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	
	// Trasformiamo la risposta res da struct LoginResponse a JSON e la inviamo al client
	_ = json.NewEncoder(w).Encode(res)
}