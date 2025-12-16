package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrForbidden = errors.New("user is not a member of this conversation")
var ErrBadRequest = errors.New("invalid request data")
var ErrAlreadyMember = errors.New("user is already a member")
var ErrUsernameTaken = errors.New("username already taken")

// Interfaccia Database
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

	// Modificata: accetta text e photoUrl opzionali
	SendMessage(senderId string, convId string, text string, photoUrl string, replyToMsgId *string) (Message, error)

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
	c *sql.DB
}

func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Users Table
	sqlStmt := `CREATE TABLE IF NOT EXISTS users (
		id TEXT NOT NULL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		photoUrl TEXT
	);`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error creating database structure: %w", err)
	}

	// Conversations Table
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

	// Members Table
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

	// Messages Table (AGGIORNATA)
	// Rimuoviamo 'content' e 'contentType'
	// Aggiungiamo 'text' e 'photoUrl'
	sqlStmt = `CREATE TABLE IF NOT EXISTS messages (
		id TEXT NOT NULL PRIMARY KEY,
		conversationId TEXT NOT NULL,
		senderId TEXT NOT NULL,
		
		text TEXT,       -- Può essere null/empty se c'è solo foto
		photoUrl TEXT,   -- Può essere null/empty se c'è solo testo

		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		replyToMsgId TEXT,
		status TEXT NOT NULL DEFAULT 'sent',
		forwardedFromMsgId TEXT,
		FOREIGN KEY (conversationId) REFERENCES conversations(id) ON DELETE CASCADE,
		FOREIGN KEY (senderId) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (replyToMsgId) REFERENCES messages(id) ON DELETE SET NULL,
		FOREIGN KEY (forwardedFromMsgId) REFERENCES messages(id) ON DELETE SET NULL
	);`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error creating messages table: %w", err)
	}

	// Reactions Table
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

	return &appdbimpl{c: db}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
