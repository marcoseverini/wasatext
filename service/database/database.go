package database

import (
	"database/sql" // Libreria per parlare con SQL
	"errors"       // Libreria per gestire gli errori
	"fmt"          // Libreria per formattare gli errori
)

var ErrForbidden = errors.New("user is not a member of this conversation")
var ErrBadRequest = errors.New("invalid request data")
var ErrAlreadyMember = errors.New("user is already a member")
var ErrUsernameTaken = errors.New("username already taken")

// Interfaccia per comunicare con il database
type AppDatabase interface {
	Ping() error
	DoLogin(username string) (User, error)
	GetUserByID(userID string) (User, error)
	GetUserByName(username string) (User, error)
	CreateUser(username string) (User, error)
	SetMyUserName(userID string, newUsername string) (User, error)
	SetMyPhoto(userID string, photoURL string) (User, error)
	SearchUsers(username string) ([]User, error)
	CheckUserExists(userID string) (bool, error)
	StartConversation(requestingUserID string, targetUserID string) (string, error)
	GetConversationDetails(conversationID string, requestingUserID string) (Conversation, error)
	GetConversationSummaries(userID string) ([]ConversationSummary, error)
	SendMessage(senderId string, convId string, content string, contentType string, replyToMsgId *string) (Message, error)
	DeleteMessage(requestingUserID string, messageID string) error
	ForwardMessage(requestingUserID string, targetConvId string, originalMessageId string) (Message, error)
	AddReaction(requestingUserID string, messageID string, emoji string) (Reaction, error)
	RemoveReaction(requestingUserID string, reactionID string, messageID string) error
	CreateGroup(requestingUserID string, groupName string, memberIds []string) (string, error)
	SetGroupName(requestingUserID string, convId string, newName string) error
	SetGroupPhoto(requestingUserID string, convId string, newPhotoURL string) error
	AddGroupMember(requestingUserID string, convId string, targetUserID string) error
	LeaveGroup(requestingUserID string, convId string) error
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

	// Tabella Conversazioni
	sqlStmt = `CREATE TABLE IF NOT EXISTS conversations (
		id TEXT NOT NULL PRIMARY KEY,
		name TEXT,
		photoUrl TEXT,
		isGroup INTEGER NOT NULL DEFAULT 0
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error creating conversations table: %w", err)
	}

	// Tabella Membri Conversazione
	sqlStmt = `CREATE TABLE IF NOT EXISTS conversation_members (
		conversationId TEXT NOT NULL,
		userId TEXT NOT NULL,
		PRIMARY KEY (conversationId, userId),
		FOREIGN KEY (conversationId) REFERENCES conversations(id) ON DELETE CASCADE,
		FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error creating conversation_members table: %w", err)
	}

	// Tabella Messaggi
	sqlStmt = `CREATE TABLE IF NOT EXISTS messages (
		id TEXT NOT NULL PRIMARY KEY,
		conversationId TEXT NOT NULL,
		senderId TEXT NOT NULL,
		content TEXT NOT NULL,
		contentType TEXT NOT NULL DEFAULT 'text',
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		replyToMsgId TEXT,
		status TEXT NOT NULL DEFAULT 'sent',
		forwardedFromMsgId TEXT, -- AGGIUNGI QUESTA
		FOREIGN KEY (conversationId) REFERENCES conversations(id) ON DELETE CASCADE,
		FOREIGN KEY (senderId) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (replyToMsgId) REFERENCES messages(id) ON DELETE SET NULL,
		FOREIGN KEY (forwardedFromMsgId) REFERENCES messages(id) ON DELETE SET NULL -- AGGIUNGI QUESTA
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error creating messages table: %w", err)
	}

	// Tabella Reazioni
	sqlStmt = `CREATE TABLE IF NOT EXISTS reactions (
		id TEXT NOT NULL PRIMARY KEY,
		messageId TEXT NOT NULL,
		userId TEXT NOT NULL,
		emoji TEXT NOT NULL,
		UNIQUE (messageId, userId, emoji),
		FOREIGN KEY (messageId) REFERENCES messages(id) ON DELETE CASCADE,
		FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error creating reactions table: %w", err)
	}

	// Ritorna un puntatore all'implementazione concreta del database (appdbimpl)
	// che soddisfa l'interfaccia (AppDatabase) ed un errore nullo (nil) per segnalare il successo
	return &appdbimpl{c: db}, nil
}

func (db *appdbimpl) Ping() error {
	// Inoltra il comando Ping alla connessione con il database per verificare che sia attiva
	return db.c.Ping()
}
