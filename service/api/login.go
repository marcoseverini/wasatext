package api

import (
	"encoding/json" // Libreria per codificare/decodificare JSON
	"net/http"      // Libreria per gestire HTTP

	"github.com/julienschmidt/httprouter" // Router HTTP di terze parti
)

// POST /session
func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	var req DoLoginRequest // components/schemas/DoLoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	if err := req.Username.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	user, err := rt.db.DoLogin(string(req.Username))
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	res := DoLoginResponse{ // components/schemas/DoLoginResponse
		Identifier: UserID(user.ID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(res)
}
