package database

import (
	"database/sql" // Libreria per parlare con SQL
	"errors" // Libreria per gestire gli errori
	"fmt" // Libreria per formattare gli errori
	"github.com/mattn/go-sqlite3" // Driver SQLite per Go
	"github.com/google/uuid" // Pacchetto per generare ID univoci
)

// Interfaccia per comunicare con il database
type AppDatabase interface { 
	Ping() error 
	DoLogin(username string) (User, error) 
	GetUserByID(userID string) (User, error)
	GetUserByName(username string) (User, error)
	CreateUser(username string) (User, error)
	SetMyUsername(userID string, newUsername string) (User, error)
	SetMyPhoto(userID string, photoURL string) (User, error) 
}

type appdbimpl struct {
	c *sql.DB // Connessione con il Database
}

func New(db *sql.DB) (AppDatabase, error) { 
	// Prende in input una connessione al database, restituisce un'istanza di AppDatabase

	if db == nil { // Verifica che la connessione al database non sia nulla
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Comando SQL per creare la tabella utenti, se non esiste
	sqlStmt := `CREATE TABLE IF NOT EXISTS users (
		id TEXT NOT NULL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		photoUrl TEXT
	);`
	_, err := db.Exec(sqlStmt) // Esegue il comando SQL sul database

	if err != nil {
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	// Ritorna un puntatore all'implementazione concreta del database (appdbimpl) 
	// che soddisfa l'interfaccia (AppDatabase) ed un errore nullo (nil) per segnalare il successo
	return &appdbimpl{c: db}, nil 
}

func (db *appdbimpl) Ping() error {
	// Inoltra il comando Ping alla connessione con il database per verificare che sia attiva
	return db.c.Ping() 
}

func (db *appdbimpl) GetUserByName(username string) (User, error) {
	// Prende un nome utente e restituisce l'utente corrispondente dal database
	// Se l'utente non esiste, restituisce un errore 

	var user User

	err := db.c.QueryRow(`SELECT id, username, photoUrl FROM users WHERE username = ?`, username).
		Scan(&user.ID, &user.Username, &user.PhotoURL)

	if err != nil {
		return user, err
	}
	return user, nil
}

func (db *appdbimpl) GetUserByID(userID string) (User, error) {
	// Prende un userID e restituisce l'utente corrispondente dal database
	// Se l'utente non esiste, restituisce un errore

    var user User
    err := db.c.QueryRow(`SELECT id, username, photoUrl FROM users WHERE id = ?`, userID).
        Scan(&user.ID, &user.Username, &user.PhotoURL) 

    if err != nil {
        // Se QueryRow non trova l'utente, restituisce sql.ErrNoRows.
        // Lo restituiamo così com'è. Altrimenti, è un altro errore SQL.
        return User{}, err // Restituisce struct vuota e l'errore
    }

    // Utente trovato, restituisci l'utente completo e nessun errore
    return user, nil
}

func (db *appdbimpl) CreateUser(username string) (User, error) {
	// Prende un nome utente e crea un nuovo utente nel database 
	// Restituisce l'utente creato o un errore
	
	newID := uuid.New().String() // Genera un nuovo ID univoco per l'utente

	user := User{ // Crea una struttura User con l'ID e il nome utente
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

	user, err := db.GetUserByName(username) // Prova a ottenere l'utente dal database

	if err == nil { // Se l'utente esiste, lo restituisce
		return user, nil
	}

	if errors.Is(err, sql.ErrNoRows) { // Se l'utente non esiste, lo crea
		return db.CreateUser(username)
	}

	return user, fmt.Errorf("error during login process: %w", err) // Altri errori
}

// Errore specifico che restituiamo quando si viola il vincolo UNIQUE.
var ErrUsernameTaken = errors.New("username already taken")

func (db *appdbimpl) SetMyUsername(userID string, newUsername string) (User, error) {
	// Prende l'ID utente dell'utente che vuole cambiare nome e il nuovo nome desiderato
	// Se l'utente non esiste, restituisce un errore
	// Se il nuovo nome è già in uso, restituisce un errore
	// Altrimenti aggiorna il nome utente e restituisce l'utente aggiornato

	// Comando SQL per aggiornare il nome utente
    sqlStmt := `UPDATE users SET username = ? WHERE id = ?`
    _, err := db.c.Exec(sqlStmt, newUsername, userID) // Eseguiamo l'aggiornamento

    if err != nil {
        // Controlliamo se l'errore è dovuto al vincolo UNIQUE
        var sqliteErr *sqlite3.Error
        if errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			// Restituiamo User{} (struct vuota) e l'errore specifico
			return User{}, ErrUsernameTaken
		}
        // Altrimenti, è un altro errore SQL
        return User{}, fmt.Errorf("error updating username: %w", err)
    }

    // Se l'aggiornamento è andato a buon fine, recuperiamo i dati aggiornati
	updatedUser, err := db.GetUserByID(userID)
    if err != nil {
        return User{}, fmt.Errorf("error fetching updated user data after username update: %w", err)
    }

    return updatedUser, nil // Restituisci l'utente aggiornato e nessun errore
}


func (db *appdbimpl) SetMyPhoto(userID string, photoURL string) (User, error) {
	// Prende l'ID utente dell'utente che vuole cambiare la foto profilo e la nuova URL desiderata

	// Comando SQL per aggiornare la foto profilo
	sqlStmt := `UPDATE users SET photoUrl = ? WHERE id = ?` 
	_, err := db.c.Exec(sqlStmt, photoURL, userID)
	if err != nil {
		return User{}, fmt.Errorf("error updating user profile photo: %w", err)
	}

	// Se l'aggiornamento è andato a buon fine, recuperiamo i dati aggiornati
	updatedUser, err := db.GetUserByID(userID)
    if err != nil {
        // Se non riusciamo a leggere l'utente appena aggiornato, c'è un problema serio.
        return User{}, fmt.Errorf("error fetching updated user data after photo update: %w", err)
    }

    return updatedUser, nil // Restituisci l'utente completo e aggiornato
}