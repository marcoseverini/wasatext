package api

import (
	"encoding/json"
	"errors" // Per usare errors.Is
	"net/http"
	"wasatext/service/database" // Importa per ErrUsernameTaken

	"github.com/julienschmidt/httprouter"
)

// setMyUserName è l'handler per PUT /settings/username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// Otteniamo l'ID utente dall'autenticazione (simulata per ora)
	// In futuro, questo ID verrà estratto dal token Bearer da un middleware.
	// Per ora, lo leggiamo da un header fittizio per poter testare.
	// Cambieremo questa parte quando implementeremo l'autenticazione.

	userID := r.Header.Get("X-User-ID") // Header fittizio per il test
	if userID == "" {
		rt.sendErrorResponse(w, http.StatusUnauthorized, "Autenticazione richiesta (simulata tramite X-User-ID)")
		return
	}

	// leggiamo la richiesta JSON
	// e la trasformiamo in una struct SetUsernameRequest
	var req SetUsernameRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido: "+err.Error())
		return
	}

	// Verifichiamo che il nuovo nome utente rispetti le regole
	// (minimo 3 caratteri, massimo 16)
	newUsername := req.Username
	if len(newUsername) < 3 || len(newUsername) > 16 {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Nuovo nome utente non valido (deve essere tra 3 e 16 caratteri)")
		return
	}

	// Aggiorniamo il nome utente nel database
	updatedUser, err := rt.db.SetUsername(userID, newUsername)
	if err != nil {
		// Controlliamo se l'errore è quello specifico di "nome già preso"
		if errors.Is(err, database.ErrUsernameTaken) {
			rt.sendErrorResponse(w, http.StatusConflict, "Username già in uso") // Risposta 409
		} else {
			// Altro errore del database
			rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore interno durante l'aggiornamento: "+err.Error())
		}
		return
	}

	// Inviamo la risposta di successo
	w.WriteHeader(http.StatusOK) // Codice 200 OK
	w.Header().Set("Content-Type", "application/json")
	// Rispondiamo con l'intera struct User aggiornata (che include ID e nuovo Username)
	_ = json.NewEncoder(w).Encode(updatedUser)
}