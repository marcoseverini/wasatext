package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url" // Importa 'net/url' per la validazione

	"github.com/julienschmidt/httprouter"
	"github.com/marcoseverini/wasatext/service/database" // Importa il package database
)

// NewMessageRequest modella il body JSON per inviare un messaggio
type NewMessageRequest struct {
	Text     *string `json:"text,omitempty"`
	PhotoURL *string `json:"photoUrl,omitempty"`
	ReplyTo  *string `json:"replyToMsgId,omitempty"`
}

// sendMessage (POST /conversations/{convId}/messages)
func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Prendi l'ID dell'utente che fa la richiesta (dal middleware)
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	// 2. Prendi l'ID della conversazione dai parametri URL
	convId := ps.ByName("convId")
	if convId == "" {
		rt.sendErrorResponse(w, http.StatusBadRequest, "ID conversazione mancante.")
		return
	}

	// 3. Decodifica il body JSON
	var req NewMessageRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "JSON non valido: "+err.Error())
		return
	}

	// 4. Valida la logica 'oneOf' (testo O foto)
	var content, contentType string

	hasText := req.Text != nil && *req.Text != ""
	hasPhoto := req.PhotoURL != nil && *req.PhotoURL != ""

	if !hasText && !hasPhoto {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Messaggio vuoto. 'text' o 'photoUrl' è richiesto.")
		return
	}
	if hasText && hasPhoto {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Non puoi inviare 'text' e 'photoUrl' contemporaneamente.")
		return
	}

	// 5. Imposta contenuto e tipo, e valida l'URL
	if hasText {
		content = *req.Text
		contentType = "text"
		// (Aggiungi validazione min/max length se vuoi)
	} else {
		// Valida l'URL della foto
		if _, err := url.ParseRequestURI(*req.PhotoURL); err != nil {
			rt.sendErrorResponse(w, http.StatusBadRequest, "URL della foto non valido.")
			return
		}
		content = *req.PhotoURL
		contentType = "photo"
	}

	// 6. Chiama il database
	newMessage, err := rt.db.SendMessage(requestingUserID, convId, content, contentType, req.ReplyTo)
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Non sei membro di questa conversazione.")
			return
		}
		if errors.Is(err, database.ErrBadRequest) {
			rt.sendErrorResponse(w, http.StatusBadRequest, "Messaggio a cui rispondere non valido o non trovato.")
			return
		}
		// Errore generico (include il 404 sulla conversazione, che il DB non gestisce)
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore invio messaggio: "+err.Error())
		return
	}

	// 7. Successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	_ = json.NewEncoder(w).Encode(newMessage)
}


// getConversation (GET /conversations/{convId})
// Questo ora è FACILISSIMO
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// 1. Prendi l'ID dell'utente (dal middleware)
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	// 2. Prendi l'ID della conversazione dai parametri URL
	convId := ps.ByName("convId")
	if convId == "" {
		rt.sendErrorResponse(w, http.StatusBadRequest, "ID conversazione mancante.")
		return
	}
	
	// 3. Chiamiamo la funzione DB che abbiamo già scritto!
	conversationDetails, err := rt.db.GetConversationDetails(convId, requestingUserID)
	if err != nil {
		// GetConversationDetails gestisce già il controllo 403/404 (restituisce errore)
		// Dobbiamo solo mappare l'errore
		// NOTA: Dovremmo migliorare il DB per distinguere 403 da 404
		rt.sendErrorResponse(w, http.StatusNotFound, "Conversazione non trovata o accesso negato.")
		return
	}

	// 4. Successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200
	_ = json.NewEncoder(w).Encode(conversationDetails)
}