package database

import (
	"database/sql"
	"fmt"
)

func (db *appdbimpl) GetUserByID(userID string) (User, error) {
	// Prende un userID e restituisce l'utente corrispondente dal database
	// Se l'utente non esiste, restituisce un errore

	var user User // components/schemas/User

	var nullablePhotoURL sql.NullString // Variabile per gestire il campo photoUrl che può essere NULL

	err := db.c.QueryRow(`SELECT id, username, photoUrl FROM users WHERE id = ?`, userID).
		Scan(&user.ID, &user.Username, &nullablePhotoURL)

	if err != nil {
		// Se QueryRow non trova l'utente, restituisce sql.ErrNoRows.
		// Lo restituisce così com'è. Altrimenti, è un altro errore SQL.
		return User{}, err // Restituisce struct vuota e l'errore
	}

	if nullablePhotoURL.Valid { // Controlla se photoUrl non è NULL
		user.PhotoURL = nullablePhotoURL.String
	}

	// Utente trovato, restituisce l'utente completo e nessun errore
	return user, nil
}

func (db *appdbimpl) CheckUserExists(userID string) (bool, error) {
	// Verifica se un utente esiste nel DB.

	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking user existence: %w", err)
	}
	return exists, nil
}

// Cerca gli utenti il cui nome utente contiene la stringa fornita.
func (db *appdbimpl) SearchUsers(username string) ([]User, error) {

	// Lista per contenere i risultati
	var users []User // Lista di components/schemas/User

	// Stringa di ricerca con i caratteri jolly
	searchQuery := "%" + username + "%"

	// Esegue la query per trovare tutti gli utenti che corrispondono
	rows, err := db.c.Query(`SELECT id, username, photoUrl FROM users WHERE username LIKE ?`, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("error searching users: %w", err)
	}
	defer rows.Close()

	// Itera su ogni riga (utente) trovata
	for rows.Next() {
		var user User
		var nullablePhotoURL sql.NullString

		// Scansiona i dati della riga nella struct User
		if err := rows.Scan(&user.ID, &user.Username, &nullablePhotoURL); err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}

		if nullablePhotoURL.Valid {
			user.PhotoURL = nullablePhotoURL.String
		}

		// Aggiunge l'utente alla lista dei risultati
		users = append(users, user)
	}

	// Controlla se ci sono stati errori durante l'iterazione
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	// Se nessun utente è stato trovato, 'users' sarà una lista vuota (non nil)
	return users, nil
}
