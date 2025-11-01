package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/julienschmidt/httprouter"
	"github.com/marcoseverini/wasatext/service/database"
)

// --- STRUCT PER JSON BODY ---

type NewGroupRequest struct {
	GroupName string   `json:"groupName"`
	MemberIds []string `json:"memberIds"`
}

type SetGroupNameRequest struct {
	Name string `json:"name"`
}

type AddMemberRequest struct {
	UserID string `json:"userId"`
}

// --- HANDLERS ---

// createGroup (POST /groups)
func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// 1. Prendi l'ID utente
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	// 2. Decodifica body
	var req NewGroupRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido.")
		return
	}

	// 3. Validazione
	if req.GroupName == "" || len(req.MemberIds) == 0 {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Nome gruppo e lista membri sono richiesti.")
		return
	}
	// (Qui potresti aggiungere un controllo per ID duplicati o per l'ID del creatore)

	// 4. Chiama il DB
	convID, err := rt.db.CreateGroup(requestingUserID, req.GroupName, req.MemberIds)
	if err != nil {
		// Potrebbe fallire se un ID membro non esiste (FOREIGN KEY)
		rt.sendErrorResponse(w, http.StatusBadRequest, "Dati non validi: "+err.Error())
		return
	}

	// 5. Restituisci la conversazione completa (come da YAML)
	conversationDetails, err := rt.db.GetConversationDetails(convID, requestingUserID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare i dettagli del gruppo.")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	_ = json.NewEncoder(w).Encode(conversationDetails)
}

// setGroupName (PUT /conversations/{convId}/name)
func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Prendi ID utente e ID conversazione
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	convId := ps.ByName("convId")

	// 2. Decodifica body
	var req SetGroupNameRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido o nome mancante.")
		return
	}

	// 3. Chiama il DB
	err = rt.db.SetGroupName(requestingUserID, convId, req.Name)
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questo gruppo.")
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non è un gruppo.")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Conversazione non trovata.")
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'aggiornamento.")
		return
	}

	// 4. Restituisci la conversazione aggiornata
	conversationDetails, err := rt.db.GetConversationDetails(convId, requestingUserID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare i dettagli del gruppo.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200
	_ = json.NewEncoder(w).Encode(conversationDetails)
}

// setGroupPhoto (PUT /conversations/{convId}/photo)
func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Prendi ID utente e ID conversazione
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	convId := ps.ByName("convId")

	// 2. Decodifica body (Usa la struct SetPhotoRequest di 'settings.go')
	var req SetPhotoRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido.")
		return
	}
	if _, err = url.ParseRequestURI(req.PhotoURL); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "URL non valido.")
		return
	}

	// 3. Chiama il DB
	err = rt.db.SetGroupPhoto(requestingUserID, convId, req.PhotoURL)
	if err != nil {
		// (Gestione errori identica a setGroupName)
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questo gruppo.")
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non è un gruppo.")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Conversazione non trovata.")
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'aggiornamento.")
		return
	}

	// 4. Restituisci la conversazione aggiornata
	conversationDetails, err := rt.db.GetConversationDetails(convId, requestingUserID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare i dettagli del gruppo.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200
	_ = json.NewEncoder(w).Encode(conversationDetails)
}

// addToGroup (POST /conversations/{convId}/members)
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Prendi ID utente e ID conversazione
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	convId := ps.ByName("convId")

	// 2. Decodifica body
	var req AddMemberRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido o 'userId' mancante.")
		return
	}

	// 3. Chiama il DB
	err = rt.db.AddGroupMember(requestingUserID, convId, req.UserID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrForbidden):
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questo gruppo.")
		case errors.Is(err, database.ErrBadRequest):
			rt.sendErrorResponse(w, http.StatusForbidden, "Non è un gruppo.")
		case errors.Is(err, sql.ErrNoRows):
			rt.sendErrorResponse(w, http.StatusNotFound, "Gruppo o utente da aggiungere non trovato.")
		case errors.Is(err, database.ErrAlreadyMember):
			rt.sendErrorResponse(w, http.StatusConflict, "L'utente è già membro del gruppo.")
		default:
			rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'aggiunta del membro.")
		}
		return // Ritorna dopo aver gestito l'errore
	}

	// 4. Successo
	w.WriteHeader(http.StatusNoContent) // 204
}

// leaveGroup (DELETE /conversations/{convId}/members/me)
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Prendi ID utente e ID conversazione
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}
	convId := ps.ByName("convId")

	// 2. Chiama il DB
	err = rt.db.LeaveGroup(requestingUserID, convId)
	if err != nil {
		if errors.Is(err, database.ErrForbidden) || errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questo gruppo (o non è un gruppo).")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Gruppo non trovato.")
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'abbandono del gruppo.")
		return
	}

	// 3. Successo
	w.WriteHeader(http.StatusNoContent) // 204
}
