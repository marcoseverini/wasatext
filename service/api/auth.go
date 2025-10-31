package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// contextKey è un tipo personalizzato per la nostra chiave del context.
type contextKey string

const (
	// userIdentifierKey è la chiave che usiamo per salvare l'ID dell'utente nel context della richiesta.
	userIdentifierKey contextKey = "userIdentifier"
)

// Avvolge un httprouter.Handle e controlla l'autenticazione.
func (rt *_router) authMiddleware(next httprouter.Handle) httprouter.Handle {

	// Restituiamo il nuovo Handle che fa i controlli
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// Legge l'header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			rt.sendErrorResponse(w, http.StatusUnauthorized, "Autorizzazione richiesta. Token Bearer mancante.")
			return
		}

		// Controlla che inizi con "Bearer"
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			rt.sendErrorResponse(w, http.StatusUnauthorized, "Token malformato o tipo non Bearer.")
			return
		}

		// Estrae il token (che per noi è l'identifier)
		userIdentifier := headerParts[1]

		// Verifica che l'identifier esista nel database.
		user, err := rt.db.GetUserByID(userIdentifier)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// L'identifier non corrisponde a nessun utente
				rt.sendErrorResponse(w, http.StatusUnauthorized, "Token non valido.")
			} else {
				// Errore generico del database
				rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore interno durante la validazione del token.")
			}
			return
		}

		// Se l'utente è valido, alleghiamo l'ID dell'utente (user.ID) al context della richiesta.
		ctx := context.WithValue(r.Context(), userIdentifierKey, user.ID)

		// 6. Chiama l'handler finale (es. setMyUserName) con il nuovo context
		next(w, r.WithContext(ctx), ps)
	}
}

// Estrae l'ID utente dal context (che è stato inserito da authMiddleware)
func (rt *_router) getUserIdFromAuth(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(userIdentifierKey).(string)
	if !ok {
		// Questo errore significa che l'handler è stato chiamato  senza essere protetto da authMiddleware.
		return "", errors.New("ID utente non trovato nel context")
	}
	return userID, nil
}