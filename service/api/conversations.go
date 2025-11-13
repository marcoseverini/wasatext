package api

import (
	"database/sql"  // Libreria per interagire con database SQL
	"encoding/json" // Libreria per la codifica/decodifica JSON
	"errors"        // Libreria per gestire gli errori
	"net/http"      // Libreria per gestire HTTP

	"github.com/julienschmidt/httprouter"                // Router HTTP di terze parti
	"github.com/marcoseverini/wasatext/service/database" // Database
)

// POST /conversations
func (rt *_router) startConversation(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	var req UserIdRequest // components/schemas/UserIdRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Richiesta non valida. 'userId' mancante o malformato.") // 400 Bad Request
		return
	}

	if err := req.UserID.Validate(); err != nil {
		rt.sendErrorResponse(w, http.StatusBadRequest, err.Error()) // 400 Bad Request
		return
	}

	// Controlla che l'utente non stia cercando di iniziare una chat con se stesso
	if userID == string(req.UserID) {
		rt.sendErrorResponse(w, http.StatusBadRequest, "Non puoi iniziare una conversazione con te stesso.") // 400 Bad Request
		return
	}

	// Controlla se l'utente target esiste
	exists, err := rt.db.CheckUserExists(string(req.UserID))
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel verificare l'utente.") // 500 Internal Server Error
		return
	}
	if !exists {
		rt.sendErrorResponse(w, http.StatusNotFound, "L'utente specificato non esiste.") // 404 Not Found
		return
	}

	// Avvia (o trova) la conversazione nel DB
	convID, err := rt.db.StartConversation(userID, string(req.UserID))
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nell'avviare la conversazione.") // 500 Internal Server Error
		return
	}

	// Recupera i dettagli della conversazione appena creata e li restituisce
	conversationDetails, err := rt.db.GetConversationDetails(convID, userID) // components/schemas/Conversation
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare i dettagli della conversazione.") // 500 Internal Server Error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(conversationDetails)
}

// GET /conversations
func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	userID, err := rt.getUserIdFromAuth(r) // Autenticazione
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, err.Error()) // 500 Internal Server Error
		return
	}

	// Chiede al DB il riepilogo delle conversazioni
	summaries, err := rt.db.GetConversationSummaries(userID) // components/schemas/ConversationSummary
	if err != nil {
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare le conversazioni.") // 500 Internal Server Error
		return
	}

	response := database.ConversationList{
		Conversations: summaries,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	_ = json.NewEncoder(w).Encode(response)
}

// GET /conversations/{convId}
func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

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

	conversationDetails, err := rt.db.GetConversationDetails(string(convId), userID) // components/schemas/Conversation
	if err != nil {
		if errors.Is(err, database.ErrForbidden) {
			rt.sendErrorResponse(w, http.StatusForbidden, "Accesso negato.") // 403 Forbidden
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			rt.sendErrorResponse(w, http.StatusNotFound, "Conversazione non trovata.") // 404 Not Found
			return
		}
		rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore nel recuperare la conversazione.") // 500 Internal Server Error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	_ = json.NewEncoder(w).Encode(conversationDetails)
}
