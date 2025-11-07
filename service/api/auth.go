package api

import (
	"context"      // Libreria per gestire i context in Go
	"database/sql" // Libreria per interagire con database SQL
	"errors"       // Libreria per gestire gli errori
	"net/http"     // Libreria per gestire HTTP
	"strings"      // Libreria per manipolare stringhe

	"github.com/julienschmidt/httprouter" // Router HTTP di terze parti
)

type contextKey string // Tipo personalizzato per le chiavi del context

const (
	// Chiave usata per salvare l'ID dell'utente nel context della richiesta.
	userIdentifierKey contextKey = "userIdentifier"
)

// Avvolge un httprouter.Handle e controlla l'autenticazione.
func (rt *_router) authMiddleware(next httprouter.Handle) httprouter.Handle {

	// Restituisce una nuova funzione che implementa il middleware
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// Legge l'header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			rt.sendErrorResponse(w, http.StatusUnauthorized, "Autorizzazione richiesta. Token Bearer mancante.") // 401 Unauthorized
			return
		}

		// Controlla che inizi con "Bearer"
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			rt.sendErrorResponse(w, http.StatusUnauthorized, "Token malformato o tipo non Bearer.") // 401 Unauthorized
			return
		}

		// Estrae il token (l'identifier)
		userIdentifier := headerParts[1]

		// Verifica che l'identifier esista nel database.
		user, err := rt.db.GetUserByID(userIdentifier)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// L'identifier non corrisponde a nessun utente
				rt.sendErrorResponse(w, http.StatusUnauthorized, "Token non valido.") // 401 Unauthorized
			} else {
				rt.sendErrorResponse(w, http.StatusInternalServerError, "Errore interno durante la validazione del token.") // 500 Internal Server Error
			}
			return
		}

		// Se l'utente è valido, allega l'ID dell'utente al context della richiesta.
		ctx := context.WithValue(r.Context(), userIdentifierKey, user.ID)

		// Chiama l'handler finale (es. setMyUserName) con il nuovo context
		next(w, r.WithContext(ctx), ps)
	}
}

// Estrae l'ID utente dal context (che è stato inserito da authMiddleware)
func (rt *_router) getUserIdFromAuth(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(userIdentifierKey).(string)
	if !ok {
		return "", errors.New("ID utente non trovato nel context")
	}
	return userID, nil
}
