package api

import (
	"encoding/json"                       // Libreria per codificare/decodificare JSON
	"github.com/julienschmidt/httprouter" // router HTTP di terze parti
	"net/http"                            // Strumenti per gestire l'HTTP
)

// Handler per l'endpoint POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// r è la richiesta JSON in entrata
	// w è la risposta JSON in uscita
	// _ sono i parametri dell'URL

	// Leggiamo la richiesta JSON e la trasformiamo in una struct LoginRequest
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido: "+err.Error())
		return
	}

	// Controlliamo che il nome rispetti le regole del nostro api.yaml (min: 3, max: 16)
	if len(req.Name) < 3 || len(req.Name) > 16 {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Nome utente non valido (deve essere tra 3 e 16 caratteri)")
		return
	}

	// Passiamo il nome utente al database per fare il login
	// Riceviamo indietro l'utente (esistente o appena creato), oppure un errore
	user, err := rt.db.DoLogin(req.Name)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore interno del server: "+err.Error())
		return
	}

	// Prepariamo la risposta JSON (identificatore dell'utente) da inviare al client
	res := LoginResponse{
		Identifier: user.ID,
	}

	// Impostiamo il codice di stato a "201 Created"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	// Trasformiamo la risposta res da struct LoginResponse a JSON e la inviamo al client
	_ = json.NewEncoder(w).Encode(res)
}
