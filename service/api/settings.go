package api

import (
	"encoding/json"                                      // Libreria per codificare/decodificare JSON
	"errors"                                             // Libreria per gestire gli errori
	"github.com/julienschmidt/httprouter"                // router HTTP di terze parti
	"github.com/marcoseverini/wasatext/service/database" // Il nostro database
	"net/http"                                           // Strumenti per gestire l'HTTP
	"net/url"                                            // Libreria per gestire gli URL
)

// Handler per PUT /settings/username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// r è la richiesta JSON in entrata
	// w è la risposta JSON in uscita
	// _ sono i parametri dell'URL

	// Il middleware 'authMiddleware' ha già verificato il token
	// e ha messo l'ID utente nel context. Lo recuperiamo.
	userID, ok := r.Context().Value(userIdentifierKey).(string)
	if !ok {
		// Questo non dovrebbe mai accadere se il middleware è applicato correttamente
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore interno: ID utente non trovato nel context")
		return
	}

	// Leggiamo la richiesta JSON e la trasformiamo in una struct SetUsernameRequest
	var req SetUsernameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido: "+err.Error())
		return
	}

	// Controlliamo che il nome rispetti le regole del nostro api.yaml (min: 3, max: 16)
	newUsername := req.Username
	if len(newUsername) < 3 || len(newUsername) > 16 {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Nuovo nome utente non valido (deve essere tra 3 e 16 caratteri)")
		return
	}

	// Aggiorniamo il nome utente nel database
	updatedUser, err := rt.db.SetMyUsername(userID, newUsername)
	if err != nil {
		// Controlliamo se l'errore è dovuto a username già in uso
		if errors.Is(err, database.ErrUsernameTaken) {
			rt.sendErrorResponse(w, http.StatusConflict, "Username già in uso")
		} else {
			// Altro errore del database
			rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore interno durante l'aggiornamento: "+err.Error())
		}
		return
	}

	// Inviamo la risposta di successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Codice 200 OK
	// Rispondiamo con l'intera struct User aggiornata (che include ID e nuovo Username)
	_ = json.NewEncoder(w).Encode(updatedUser)
}

// Handler per PUT /settings/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// r è la richiesta JSON in entrata
	// w è la risposta JSON in uscita
	// _ sono i parametri dell'URL

	// Il middleware 'authMiddleware' ha già verificato il token
	// e ha messo l'ID utente nel context. Lo recuperiamo.
	userID, ok := r.Context().Value(userIdentifierKey).(string)
	if !ok {
		// Questo non dovrebbe mai accadere se il middleware è applicato correttamente
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore interno: ID utente non trovato nel context")
		return
	}

	// Leggi il JSON body nella struct corretta
	var req SetPhotoRequest // Usa la struct corretta
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido: "+err.Error())
		return
	}

	photoURL := req.PhotoURL               // Estrai l'URL della foto dalla richiesta
	_, err = url.ParseRequestURI(photoURL) // Verifica che sia un URL valido
	if err != nil || photoURL == "" {      // Verifica che non sia vuoto
		rt.sendErrorResponse(w, http.StatusBadRequest, "URL della foto non valido: "+err.Error())
		return
	}

	// Chiama il database per aggiornare la foto
	updatedUser, err := rt.db.SetMyPhoto(userID, photoURL)
	if err != nil {
		// Gestione errore generico del database
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore interno durante l'aggiornamento della foto: "+err.Error())
		return
	}

	// Invia la risposta di successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(updatedUser)
}
