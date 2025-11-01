package api

import (
	"encoding/json"                       // Libreria per codificare e decodificare JSON
	"github.com/julienschmidt/httprouter" // Router HTTP di terze parti
	"github.com/marcoseverini/wasatext/service/database"
	"net/http" // Libreria per gestire richieste e risposte HTTP
)

// searchUsers è l'handler per GET /users
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Estrai il parametro 'username' dalla query string (es. /users?username=Mar)
	username := r.URL.Query().Get("username")

	// Validazione (come da api.yaml: minLength: 1, maxLength: 16)
	if len(username) < 1 || len(username) > 16 {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Il parametro 'username' deve essere tra 1 e 16 caratteri.")
		return
	}

	// Esegui la ricerca nel database
	// Per questo endpoint, l'utente autenticato e quello cercato sono diversi.
	// Assicuriamoci che l'utente non stia cercando se stesso (o gestiamo questo caso)

	// Prendiamo l'ID dell'utente che fa la richiesta (dal middleware)
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		// Se c'è un errore qui, significa che il token non è valido o mancante.
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Esegui la ricerca nel database
	users, err := rt.db.SearchUsers(username)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante la ricerca degli utenti.")
		return
	}

	// Filtra l'utente che fa la richiesta dai risultati
	var filteredUsers []database.User
	for _, user := range users {
		if user.ID != requestingUserID {
			filteredUsers = append(filteredUsers, user)
		}
	}

	// Restituisci la lista (filtrata)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Se non ci sono risultati, filteredUsers sarà una lista vuota `[]` che json.NewEncoder gestirà correttamente (come richiesto dallo YAML).
	_ = json.NewEncoder(w).Encode(filteredUsers)
}
