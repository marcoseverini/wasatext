package database

import (
	"database/sql" // Libreria per parlare con SQL
	"errors" // Libreria per gestire gli errori
	"fmt" // Libreria per formattare gli errori

	"github.com/google/uuid" // Pacchetto per generare ID univoci
)

type AppDatabase interface { // Interfaccia per comunicare con il database
	Ping() error 
	DoLogin(username string) (User, error) 
	GetUserByName(username string) (User, error)
	CreateUser(username string) (User, error)
}

type appdbimpl struct {
	c *sql.DB // Connessione con il Database
}

func New(db *sql.DB) (AppDatabase, error) { 
	// Prende in input una connessione al database, restituisce un'istanza di AppDatabase
	// Viene chiamato all'avvio del server 

	if db == nil { // Verifica che la connessione al database non sia nulla
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Comando SQL per creare la tabella utenti, se non esiste
	sqlStmt := `CREATE TABLE IF NOT EXISTS users (
		id TEXT NOT NULL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE
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
	return db.c.Ping() // inoltra il comando Ping alla connessione con il database
}

func (db *appdbimpl) GetUserByName(username string) (User, error) {
	// Prende un nome utente e restituisce l'utente corrispondente dal database
	// Se l'utente non esiste, restituisce un errore 

	var user User

	err := db.c.QueryRow(`SELECT id, username FROM users WHERE username = ?`, username).
		Scan(&user.ID, &user.Username)

	if err != nil {
		return user, err
	}
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

	// Comando SQL per inserire un nuovo utente nel database
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

