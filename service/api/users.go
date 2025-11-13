package api

import (
	"encoding/json" // Libreria per codificare/decodificare JSON
	"net/http"      // Libreria per gestire HTTP

	"github.com/julienschmidt/httprouter"                // Router HTTP di terze parti
	"github.com/marcoseverini/wasatext/service/database" // Database
)

// GET /users
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	searchQuery := r.URL.Query().Get("username") // components/parameters/SearchUsername

	var username Username = Username(searchQuery) // components/schemas/Username
	if err := username.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	// Esegue la ricerca nel database
	users, err := rt.db.SearchUsers(string(username)) // Lista di components/schemas/User
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante la ricerca degli utenti.") // 500 Internal Server Error
		return
	}

	// Filtra l'utente che fa la richiesta dai risultati
	var filteredUsers []database.User
	for _, user := range users {
		if user.ID != userID {
			filteredUsers = append(filteredUsers, user)
		}
	}

	response := database.UserList{
		Users: filteredUsers,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	_ = json.NewEncoder(w).Encode(response)
}
