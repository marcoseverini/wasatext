package database

import (
	"database/sql" // Libreria per parlare con SQL
	"errors"       // Libreria per gestire gli errori
	"fmt"          // Libreria per formattare gli errori

	"github.com/google/uuid" // Pacchetto per generare ID univoci
)

func (db *appdbimpl) GetUserByName(username string) (User, error) {
	// Prende un nome utente e restituisce l'utente corrispondente dal database
	// Se l'utente non esiste, restituisce un errore

	var nullablePhotoURL sql.NullString // Variabile per gestire il campo photoUrl che può essere NULL

	var user User // components/schemas/User

	err := db.c.QueryRow(`SELECT id, username, photoUrl FROM users WHERE username = ?`, username).
		Scan(&user.ID, &user.Username, &nullablePhotoURL)

	if err != nil {
		return user, err
	}

	if nullablePhotoURL.Valid { // Controlla se photoUrl non è NULL
		user.PhotoURL = nullablePhotoURL.String
	}

	return user, nil
}

func (db *appdbimpl) CreateUser(username string) (User, error) {
	// Prende un nome utente e crea un nuovo utente nel database
	// Restituisce l'utente creato o un errore

	newID := uuid.New().String() // Genera un nuovo ID univoco per l'utente

	// Crea una struttura User con l'ID e il nome utente
	user := User{ // components/schemas/User
		ID:       newID,
		Username: username,
	}

	sqlStmt := `INSERT INTO users (id, username) VALUES (?, ?)`
	_, err := db.c.Exec(sqlStmt, user.ID, user.Username)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (db *appdbimpl) DoLogin(username string) (User, error) {
	// Prende un nome utente, se esiste lo restituisce
	// altrimenti crea un nuovo utente e lo restituisce

	// Prova a ottenere l'utente dal database
	user, err := db.GetUserByName(username) // components/schemas/User

	if err == nil { // Se l'utente esiste, lo restituisce
		return user, nil
	}

	if errors.Is(err, sql.ErrNoRows) { // Se l'utente non esiste, lo crea
		return db.CreateUser(username)
	}

	return user, fmt.Errorf("error during login process: %w", err) // Altri errori
}
