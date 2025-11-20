package api

import (
	"encoding/json" // Libreria per codificare/decodificare JSON
	"errors"        // Libreria per gestire gli errori
	"net/http"      // Libreria per gestire HTTP

	"github.com/julienschmidt/httprouter"                // Router HTTP di terze parti
	"github.com/marcoseverini/wasatext/service/database" // Database
)

// PUT /settings/username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var req SetMyUserNameRequest // components/schemas/SetMyUserNameRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	if err := req.Username.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	// Aggiorna il nome utente nel database
	updatedUser, err := rt.db.SetMyUserName(userID, string(req.Username)) // components/schemas/User
	if err != nil {
		if errors.Is(err, database.ErrUsernameTaken) {
			rt.sendErrorResponse(w, http.StatusConflict, "Username già in uso") // 409 Conflict
		} else {
			rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	_ = json.NewEncoder(w).Encode(updatedUser)
}

// PUT /settings/photo
func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var req SetPhotoRequest // components/schemas/SetPhotoRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	if err := req.PhotoURL.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	// Aggiorna la foto nel database
	updatedUser, err := rt.db.SetMyPhoto(userID, string(req.PhotoURL)) // components/schemas/User
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	_ = json.NewEncoder(w).Encode(updatedUser)
}
