package api

import (
	"database/sql"  // Libreria per gestire SQL
	"encoding/json" // Libreria per codificare/decodificare JSON
	"errors"        // Libreria per gestire gli errori
	"net/http"      // Libreria per gestire HTTP
	"reflect"       //

	"github.com/julienschmidt/httprouter"                // Router HTTP di terze parti
	"github.com/marcoseverini/wasatext/service/database" // Database
)

// POST /conversations/{convId}/messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var convId InternalID = InternalID(ps.ByName("convId")) // components/parameters/ConvId
	if err := convId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	var req SendMessageRequest // components/schemas/SendMessageRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	var content string
	var contentType string

	hasText := req.Text != nil && !reflect.ValueOf(req.Text).IsNil()
	hasPhoto := req.PhotoURL != nil && !reflect.ValueOf(req.PhotoURL).IsNil()

	if !hasText && !hasPhoto {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Messaggio vuoto. 'text' o 'photoUrl' è richiesto.") // 400 Bad Request
		return
	}
	if hasText && hasPhoto {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Non puoi inviare 'text' e 'photoUrl' contemporaneamente.") // 400 Bad Request
		return
	}

	if hasText {
		if err := req.Text.Validate(); err != nil {
			rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
			return
		}
		content = string(*req.Text)
		contentType = "text"
	} else {
		if err := req.PhotoURL.Validate(); err != nil {
			rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
			return
		}
		content = string(*req.PhotoURL)
		contentType = "photo"
	}

	var replyTo *string
	if req.ReplyToMsgId != "" {
		if err := req.ReplyToMsgId.Validate(); err != nil {
			rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
			return
		}
		replyToString := string(req.ReplyToMsgId)
		replyTo = &replyToString
	}

	newMessage, err := rt.db.SendMessage(userID, string(convId), content, contentType, replyTo) // components/schemas/Message
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questa conversazione.") // 403 Forbidden
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusBadRequest, "Messaggio a cui rispondere non valido o non trovato.") // 400 Bad Request
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(newMessage)
}

// POST /conversations/{convId}/forwarded_messages
func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var convId InternalID = InternalID(ps.ByName("convId")) // components/parameters/ConvId
	if err := convId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	var req ForwardMessageRequest // components/schemas/ForwardMessageRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	if err := req.OriginalMessageId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	forwardedMessage, err := rt.db.ForwardMessage(userID, string(convId), string(req.OriginalMessageId)) // components/schemas/Message
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro della conversazione di destinazione.") // 403 Forbidden
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Messaggio originale non trovato o accesso negato.") // 404 Not Found
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'inoltro del messaggio.") // 500 Internal Server Error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(forwardedMessage)
}

// DELETE /messages/{msgId}
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var msgId InternalID = InternalID(ps.ByName("msgId")) // components/parameters/MsgID
	if err := msgId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	// Chiama il database per eliminare
	err = rt.db.DeleteMessage(userID, string(msgId))
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei il mittente di questo messaggio.") // 403 Forbidden
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Messaggio non trovato.") // 404 Not Found
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'eliminazione del messaggio.") // 500 Internal Server Error
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

// POST /messages/{msgId}/reactions
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var msgId InternalID = InternalID(ps.ByName("msgId")) // components/parameters/MsgID
	if err := msgId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	var req CommentMessageRequest // components/schemas/CommentMessageRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	if err := req.Emoji.Validate(); err != nil { // components/schemas/Reaction
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	reaction, err := rt.db.AddReaction(userID, string(msgId), string(req.Emoji))
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non puoi reagire a questo messaggio (non sei membro).") // 403 Forbidden
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Messaggio non trovato.") // 404 Not Found
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusBadRequest, "Emoji non valida.") // 400 Bad Request
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore durante l'aggiunta della reazione.") // 500 Internal Server Error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(reaction)
}

// DELETE /messages/{msgId}/reactions/{reactionId}
func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var msgId InternalID = InternalID(ps.ByName("msgId")) // components/parameters/MsgID
	if err := msgId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}
	var reactionId InternalID = InternalID(ps.ByName("reactionId")) // components/parameters/ReactionID
	if err := reactionId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	err = rt.db.RemoveReaction(userID, string(reactionId), string(msgId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Reazione non trovata.") // 404 Not Found
			return
		}
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei il proprietario di questa reazione.") // 403 Forbidden
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
