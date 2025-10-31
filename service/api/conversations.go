package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// startConversation (POST /conversations)
func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// 1. Prendi l'ID dell'utente che fa la richiesta (dal middleware)
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	// 2. Leggi il body della richiesta per ottenere il targetUserID
	var req struct {
		UserID string `json:"userId"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.UserID == "" {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Richiesta non valida. 'userId' mancante o malformato.")
		return
	}
	targetUserID := req.UserID

	// 3. Controlla che l'utente non stia cercando di iniziare una chat con se stesso
	if requestingUserID == targetUserID {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Non puoi iniziare una conversazione con te stesso.")
		return
	}

	// 4. Controlla se l'utente target esiste (per l'errore 404)
	exists, err := rt.db.CheckUserExists(targetUserID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel verificare l'utente.")
		return
	}
	if !exists {
		rt.sendErrorResponse(w, http.StatusNotFound, "L'utente specificato non esiste.")
		return
	}

	// 5. Avvia (o trova) la conversazione nel DB
	convID, err := rt.db.StartConversation(requestingUserID, targetUserID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nell'avviare la conversazione.")
		return
	}

	// 6. Lo YAML richiede di restituire l'intera conversazione (201)
	// (anche se è nuova e non ha messaggi)
	conversationDetails, err := rt.db.GetConversationDetails(convID, requestingUserID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare i dettagli della conversazione.")
		return
	}

	// 7. Successo
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(conversationDetails)
}


// getMyConversations (GET /conversations)
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// 1. Prendi l'ID dell'utente che fa la richiesta (dal middleware)
	requestingUserID, err := rt.getUserIdFromAuth(r)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	// 2. Chiedi al DB i riepiloghi
	summaries, err := rt.db.GetConversationSummaries(requestingUserID)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare le conversazioni.")
		return
	}
	
	// 3. Restituisci la lista (sarà `[]` se vuota, che è corretto)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(summaries)
}