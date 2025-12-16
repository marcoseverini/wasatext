package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
	"github.com/marcoseverini/wasatext/service/database"
)

// POST /conversations/{convId}/messages
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	var convId InternalID = InternalID(ps.ByName("convId"))
	if err := convId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	var req SendMessageRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validazione: Almeno uno dei due (Text o Photo) deve esserci
	hasText := req.Text != nil && !reflect.ValueOf(req.Text).IsNil()
	hasPhoto := req.PhotoURL != nil && !reflect.ValueOf(req.PhotoURL).IsNil()

	if !hasText && !hasPhoto {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Messaggio vuoto. Inserire testo o foto.")
		return
	}

	// Recuperiamo i valori
	var textContent string
	if hasText {
		if err := req.Text.Validate(); err != nil {
			rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		textContent = string(*req.Text)
	}

	var photoUrlContent string
	if hasPhoto {
		if err := req.PhotoURL.Validate(); err != nil {
			rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		photoUrlContent = string(*req.PhotoURL)
	}

	var replyTo *string
	if req.ReplyToMsgId != "" {
		if err := req.ReplyToMsgId.Validate(); err != nil {
			rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		replyToString := string(req.ReplyToMsgId)
		replyTo = &replyToString
	}

	// Chiamata al DB aggiornata
	newMessage, err := rt.db.SendMessage(userID, string(convId), textContent, photoUrlContent, replyTo)
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questa conversazione.")
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusBadRequest, "Messaggio a cui rispondere non valido.")
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(newMessage)
}

// forwardMessage, deleteMessage e reactions restano simili al vecchio ma per completezza:

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var convId InternalID = InternalID(ps.ByName("convId"))
	if err := convId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	var req ForwardMessageRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.OriginalMessageId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	forwardedMessage, err := rt.db.ForwardMessage(userID, string(convId), string(req.OriginalMessageId))
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro della conversazione di destinazione.")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Messaggio originale non trovato.")
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore inoltro.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(forwardedMessage)
}

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var msgId InternalID = InternalID(ps.ByName("msgId"))
	if err := msgId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	err = rt.db.DeleteMessage(userID, string(msgId))
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei il mittente.")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Messaggio non trovato.")
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore eliminazione.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var msgId InternalID = InternalID(ps.ByName("msgId"))
	if err := msgId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	var req CommentMessageRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := req.Emoji.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	reaction, err := rt.db.AddReaction(userID, string(msgId), string(req.Emoji))
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non puoi reagire.")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Messaggio non trovato.")
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusBadRequest, "Emoji non valida.")
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore aggiunta reazione.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(reaction)
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	var msgId InternalID = InternalID(ps.ByName("msgId"))
	if err := msgId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	var reactionId InternalID = InternalID(ps.ByName("reactionId"))
	if err := reactionId.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	err = rt.db.RemoveReaction(userID, string(reactionId), string(msgId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Reazione non trovata.")
			return
		}
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non tua reazione.")
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
