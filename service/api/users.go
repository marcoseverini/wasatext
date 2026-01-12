package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/marcoseverini/wasatext/service/database"
)

// GET /users
func (rt *_router) searchUsers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	_, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	searchQuery := r.URL.Query().Get("username")

	var searchTerm SearchQuery = SearchQuery(searchQuery)
	if err := searchTerm.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Esegue la ricerca nel database
	users, err := rt.db.SearchUsers(string(searchTerm))
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante la ricerca degli utenti.")
		return
	}

	// Se users è nil (nessun risultato), inizializziamo una slice vuota per evitare "null" nel JSON
	if users == nil {
		users = []database.User{}
	}

	response := database.UserList{
		Users: users,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
