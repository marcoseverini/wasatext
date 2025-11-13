package api

import (
	"database/sql"  // Libreria per interagire con database SQL
	"encoding/json" // Libreria per la codifica/decodifica JSON
	"errors"        // Libreria per gestire gli errori
	"net/http"      // Libreria per gestire HTTP

	"github.com/julienschmidt/httprouter"                // Router HTTP di terze parti
	"github.com/marcoseverini/wasatext/service/database" // Database
)

// POST /groups
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var req CreateGroupRequest // components/schemas/CreateGroupRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	if err = req.GroupName.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}
	if len(req.MemberIds) == 0 {
		rt.sendErrorResponse(w, http.StatusBadRequest, "La lista dei membri non può essere vuota.") // 400 Bad Request
		return
	}

	var memberIdsStrings []string = make([]string, 0, len(req.MemberIds))
	for _, memberId := range req.MemberIds {
		if err = memberId.Validate(); err != nil {
			rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
			return
		}
		memberIdsStrings = append(memberIdsStrings, string(memberId))
	}

	convID, err := rt.db.CreateGroup(userID, string(req.GroupName), memberIdsStrings) // components/schemas/Conversation
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	conversationDetails, err := rt.db.GetConversationDetails(convID, userID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare i dettagli del gruppo.") // 500 Internal Server Error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(conversationDetails)
}

// PUT /conversations/{convId}/name
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var convId InternalID = InternalID(ps.ByName("convId")) // components/parameters/ConvID
	if err = convId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var req SetGroupNameRequest // components/schemas/SetGroupNameRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido.") // 400 Bad Request
		return
	}

	if err = req.Name.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	err = rt.db.SetGroupName(userID, string(convId), string(req.Name))
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questo gruppo.") // 403 Forbidden
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non è un gruppo.") // 403 Forbidden
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Conversazione non trovata.") // 404 Not Found
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'aggiornamento.") // 500 Internal Server Error
		return
	}

	conversationDetails, err := rt.db.GetConversationDetails(string(convId), userID) // components/schemas/Conversation
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare i dettagli del gruppo.") // 500 Internal Server Error
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	_ = json.NewEncoder(w).Encode(conversationDetails)
}

// PUT /conversations/{convId}/photo
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var convId InternalID = InternalID(ps.ByName("convId")) // components/parameters/ConvID
	if err = convId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	var req SetPhotoRequest // components/schemas/SetPhotoRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido.") // 400 Bad Request
		return
	}

	if err = req.PhotoURL.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	err = rt.db.SetGroupPhoto(userID, string(convId), string(req.PhotoURL))
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questo gruppo.") // 403 Forbidden
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non è un gruppo.") // 403 Forbidden
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Conversazione non trovata.") // 404 Not Found
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'aggiornamento.") // 500 Internal Server Error
		return
	}

	conversationDetails, err := rt.db.GetConversationDetails(string(convId), userID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare i dettagli del gruppo.") // components/schemas/Conversation
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	_ = json.NewEncoder(w).Encode(conversationDetails)
}

// POST /conversations/{convId}/members
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var convId InternalID = InternalID(ps.ByName("convId")) // components/parameters/ConvID
	if err = convId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	var req UserIdRequest // components/schemas/UserIdRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido o 'userId' mancante.") // 400 Bad Request
		return
	}

	if err = req.UserID.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	err = rt.db.AddGroupMember(userID, string(convId), string(req.UserID))
	if err != nil {
		switch {
		case errors.Is(err, database.ErrForbidden):
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questo gruppo.") // 403 Forbidden
		case errors.Is(err, database.ErrBadRequest):
			rt.sendErrorResponse(w, http.StatusForbidden, "Non è un gruppo.") // 403 Forbidden
		case errors.Is(err, sql.ErrNoRows):
			rt.sendErrorResponse(w, http.StatusNotFound, "Gruppo o utente da aggiungere non trovato.") // 404 Not Found
		case errors.Is(err, database.ErrAlreadyMember):
			rt.sendErrorResponse(w, http.StatusConflict, "L'utente è già membro del gruppo.") // 409 Conflict
		default:
			rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'aggiunta del membro.") // 500 Internal Server Error
		}
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

// DELETE /conversations/{convId}/members/me
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var convId InternalID = InternalID(ps.ByName("convId")) // components/parameters/ConvID
	if err = convId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	err = rt.db.LeaveGroup(userID, string(convId))
	if err != nil {
		if errors.Is(err, database.ErrForbidden) || errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questo gruppo (o non è un gruppo).") // 403 Forbidden
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Gruppo non trovato.") // 404 Not Found
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'abbandono del gruppo.") // 500 Internal Server Error
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
