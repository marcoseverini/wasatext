package database

import (
	"errors" // Libreria per gestire gli errori
	"fmt"    // Libreria per formattare gli errori

	"github.com/mattn/go-sqlite3" // SQLite
)

func (db *appdbimpl) SetMyUserName(userID string, newUsername string) (User, error) {
	// Prende l'ID utente dell'utente che vuole cambiare nome e il nuovo nome desiderato
	// Se l'utente non esiste, restituisce un errore
	// Se il nuovo nome è già in uso, restituisce un errore
	// Altrimenti aggiorna il nome utente e restituisce l'utente aggiornato

	// Comando SQL per aggiornare il nome utente
	sqlStmt := `UPDATE users SET username = ? WHERE id = ?`
	_, err := db.c.Exec(sqlStmt, newUsername, userID) // Esegue l'aggiornamento

	if err != nil {
		// Controlla se l'errore è dovuto al vincolo UNIQUE
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				// Restituisce User{} (struct vuota) e l'errore specifico
				return User{}, ErrUsernameTaken
			}
		}
		// Altrimenti, è un altro errore SQL
		return User{}, fmt.Errorf("error updating username: %w", err)
	}

	// Se l'aggiornamento è andato a buon fine, recupera i dati aggiornati
	updatedUser, err := db.GetUserByID(userID) // components/schemas/User
	if err != nil {
		return User{}, fmt.Errorf("error fetching updated user data after username update: %w", err)
	}

	return updatedUser, nil // Restituisce l'utente aggiornato e nessun errore
}

func (db *appdbimpl) SetMyPhoto(userID string, photoURL string) (User, error) {
	// Prende l'ID utente dell'utente che vuole cambiare la foto profilo e la nuova URL desiderata

	// Comando SQL per aggiornare la foto profilo
	sqlStmt := `UPDATE users SET photoUrl = ? WHERE id = ?`
	_, err := db.c.Exec(sqlStmt, photoURL, userID)
	if err != nil {
		return User{}, fmt.Errorf("error updating user profile photo: %w", err)
	}

	// Se l'aggiornamento è andato a buon fine, recupera i dati aggiornati
	updatedUser, err := db.GetUserByID(userID) // components/schemas/User
	if err != nil {
		return User{}, fmt.Errorf("error fetching updated user data after photo update: %w", err)
	}

	return updatedUser, nil // Restituisce l'utente completo e aggiornato
}
