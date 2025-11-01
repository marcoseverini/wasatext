package database

import (
	"database/sql"                // Libreria per parlare con SQL
	"errors"                      // Libreria per gestire gli errori
	"fmt"                         // Libreria per formattare gli errori
	"github.com/google/uuid"      // Pacchetto per generare ID univoci
	"github.com/mattn/go-sqlite3" // Driver SQLite per Go
	"time"    				// Libreria per gestire date e orari
)

var ErrForbidden = errors.New("user is not a member of this conversation")
var ErrBadRequest = errors.New("invalid request data")
var ErrAlreadyMember = errors.New("user is already a member")
 
// Interfaccia per comunicare con il database
type AppDatabase interface {
	Ping() error
	DoLogin(username string) (User, error)
	GetUserByID(userID string) (User, error)
	GetUserByName(username string) (User, error)
	CreateUser(username string) (User, error)
	SetMyUsername(userID string, newUsername string) (User, error)
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

func (db *appdbimpl) GetUserByName(username string) (User, error) {
	// Prende un nome utente e restituisce l'utente corrispondente dal database
	// Se l'utente non esiste, restituisce un errore

	var nullablePhotoURL sql.NullString // Variabile per gestire il campo photoUrl che può essere NULL

	var user User

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

func (db *appdbimpl) GetUserByID(userID string) (User, error) {
	// Prende un userID e restituisce l'utente corrispondente dal database
	// Se l'utente non esiste, restituisce un errore

	var user User

	var nullablePhotoURL sql.NullString // Variabile per gestire il campo photoUrl che può essere NULL

	err := db.c.QueryRow(`SELECT id, username, photoUrl FROM users WHERE id = ?`, userID).
		Scan(&user.ID, &user.Username, &nullablePhotoURL)

	if err != nil {
		// Se QueryRow non trova l'utente, restituisce sql.ErrNoRows.
		// Lo restituiamo così com'è. Altrimenti, è un altro errore SQL.
		return User{}, err // Restituisce struct vuota e l'errore
	}

	if nullablePhotoURL.Valid { // Controlla se photoUrl non è NULL
		user.PhotoURL = nullablePhotoURL.String
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
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				// Restituiamo User{} (struct vuota) e l'errore specifico
				return User{}, ErrUsernameTaken
			}
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

// SearchUsers cerca gli utenti il cui nome utente contiene la stringa fornita.
func (db *appdbimpl) SearchUsers(username string) ([]User, error) {
	// Lista per contenere i risultati
	var users []User

	// Costruiamo la stringa di ricerca con i caratteri jolly
	searchQuery := "%" + username + "%"

	// Eseguiamo la query per trovare tutti gli utenti che corrispondono
	rows, err := db.c.Query(`SELECT id, username, photoUrl FROM users WHERE username LIKE ?`, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("error searching users: %w", err)
	}
	// È importante chiudere 'rows' quando abbiamo finito
	defer rows.Close()

	// Iteriamo su ogni riga (utente) trovata
	for rows.Next() {
		var user User
		var nullablePhotoURL sql.NullString

		// Scansioniamo i dati della riga nella struct User
		if err := rows.Scan(&user.ID, &user.Username, &nullablePhotoURL); err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}

		if nullablePhotoURL.Valid {
			user.PhotoURL = nullablePhotoURL.String
		}
		
		// Aggiungiamo l'utente alla lista dei risultati
		users = append(users, user)
	}

	// Controlliamo se ci sono stati errori durante l'iterazione
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	// Se nessun utente è stato trovato, 'users' sarà una lista vuota (non nil)
	// Questo è corretto, la query ha avuto successo e ha restituito 0 risultati.
	return users, nil
}

// CheckUserExists verifica se un utente esiste nel DB.
func (db *appdbimpl) CheckUserExists(userID string) (bool, error) {
	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking user existence: %w", err)
	}
	return exists, nil
}

// StartConversation trova una chat 1-a-1 esistente o ne crea una nuova.
// Restituisce l'ID della conversazione.
func (db *appdbimpl) StartConversation(requestingUserID string, targetUserID string) (string, error) {
	
	// 1. Cerca una chat 1-a-1 (non di gruppo) esistente tra questi due utenti.
	// Questa query trova le conversazioni (c.id) che NON sono gruppi (c.isGroup = 0)
	// e che hanno ESATTAMENTE due membri (COUNT(m.userId) = 2)
	// e dove i due membri sono i nostri due utenti (HAVING ...).
	var existingConvID string
	query := `
		SELECT c.id
		FROM conversations c
		JOIN conversation_members m ON c.id = m.conversationId
		WHERE c.isGroup = 0
		GROUP BY c.id
		HAVING COUNT(m.userId) = 2
		   AND SUM(CASE WHEN m.userId = ? THEN 1 ELSE 0 END) = 1
		   AND SUM(CASE WHEN m.userId = ? THEN 1 ELSE 0 END) = 1;`

	err := db.c.QueryRow(query, requestingUserID, targetUserID).Scan(&existingConvID)
	
	if err == nil {
		// Trovata! Restituisci l'ID della conversazione esistente.
		return existingConvID, nil
	}
	
	if !errors.Is(err, sql.ErrNoRows) {
		// Errore SQL inaspettato
		return "", fmt.Errorf("error finding existing conversation: %w", err)
	}

	// 2. Non trovata (sql.ErrNoRows). Dobbiamo crearne una nuova.
	// Usiamo una transazione per assicurare che tutto venga creato correttamente.
	tx, err := db.c.Begin()
	if err != nil {
		return "", fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback() // Se qualcosa va storto, annulla

	// Crea la nuova conversazione
	newConvID := "conv-" + uuid.New().String()
	_, err = tx.Exec("INSERT INTO conversations (id, isGroup) VALUES (?, 0)", newConvID)
	if err != nil {
		return "", fmt.Errorf("could not create conversation: %w", err)
	}

	// Aggiungi il primo membro (l'utente che fa la richiesta)
	_, err = tx.Exec("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)", newConvID, requestingUserID)
	if err != nil {
		return "", fmt.Errorf("could not add requesting user to conversation: %w", err)
	}

	// Aggiungi il secondo membro (l'utente target)
	_, err = tx.Exec("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)", newConvID, targetUserID)
	if err != nil {
		return "", fmt.Errorf("could not add target user to conversation: %w", err)
	}

	// Tutto è andato bene, conferma la transazione
	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("could not commit transaction: %w", err)
	}

	return newConvID, nil
}

// GetConversationDetails recupera tutti i dettagli di una conversazione.
func (db *appdbimpl) GetConversationDetails(conversationID string, requestingUserID string) (Conversation, error) {
	var conversation Conversation

	// 1. Verifica che l'utente sia membro di questa conversazione
	var isMember bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?)", conversationID, requestingUserID).Scan(&isMember)
	if err != nil || !isMember {
		return conversation, fmt.Errorf("user not member or conversation not found") // Sarà un 403/404
	}

	// 2. Prendi i dettagli della conversazione
	var nullableName sql.NullString
	var nullablePhoto sql.NullString
	err = db.c.QueryRow("SELECT id, name, photoUrl, isGroup FROM conversations WHERE id = ?", conversationID).
		Scan(&conversation.ID, &nullableName, &nullablePhoto, &conversation.IsGroup)
	if err != nil {
		return conversation, fmt.Errorf("could not get conversation details: %w", err)
	}
	
	conversation.Name = nullableName.String
	conversation.PhotoURL = nullablePhoto.String

	// 3. Se non è un gruppo, il nome e la foto sono quelli dell'ALTRO utente
	if !conversation.IsGroup {
		var otherUser User
		var otherPhoto sql.NullString
		err = db.c.QueryRow(`
			SELECT u.id, u.username, u.photoUrl FROM users u
			JOIN conversation_members cm ON u.id = cm.userId
			WHERE cm.conversationId = ? AND cm.userId != ?`, conversationID, requestingUserID).
			Scan(&otherUser.ID, &otherUser.Username, &otherPhoto)
		
		if err == nil {
			conversation.Name = otherUser.Username
			conversation.PhotoURL = otherPhoto.String
		}
	}

	// 4. Prendi i membri
	rows, err := db.c.Query(`
		SELECT u.id, u.username, u.photoUrl FROM users u
		JOIN conversation_members cm ON u.id = cm.userId
		WHERE cm.conversationId = ?`, conversationID)
	if err != nil {
		return conversation, fmt.Errorf("could not get conversation members: %w", err)
	}
	defer rows.Close()

	var members []User
	for rows.Next() {
		var user User
		var photo sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &photo); err != nil {
			return conversation, fmt.Errorf("could not scan member: %w", err)
		}
		user.PhotoURL = photo.String
		members = append(members, user)
	}
	conversation.Members = members

	// 5. Prendi i messaggi (per ora, senza reazioni o 'sender' completo per semplicità)
	// LO YAML richiede il sender completo, quindi dobbiamo fare una JOIN
	msgRows, err := db.c.Query(`
		SELECT m.id, m.content, m.contentType, m.timestamp,
		       u.id as senderId, u.username as senderUsername, u.photoUrl as senderPhoto
		FROM messages m
		JOIN users u ON m.senderId = u.id
		WHERE m.conversationId = ?
		ORDER BY m.timestamp DESC`, conversationID)
	if err != nil {
		return conversation, fmt.Errorf("could not get messages: %w", err)
	}
	defer msgRows.Close()

	var messages []Message

	messageMap := make(map[string]*Message) 

	for msgRows.Next() {
		var msg Message
		var senderPhoto sql.NullString
		if err := msgRows.Scan(&msg.ID, &msg.Content, &msg.ContentType, &msg.Timestamp,
			&msg.Sender.ID, &msg.Sender.Username, &senderPhoto); err != nil {
			return conversation, fmt.Errorf("could not scan message: %w", err)
		}
		msg.Sender.PhotoURL = senderPhoto.String
		msg.Reactions = []Reaction{} // Inizializza come array vuoto (per [] non null)
		messages = append(messages, msg)
		// Aggiungi un puntatore al messaggio nella mappa
		messageMap[msg.ID] = &messages[len(messages)-1]
	}
	msgRows.Close() // Chiudi qui perché abbiamo finito con msgRows
	

	// 6. [MODIFICA] Prendi TUTTE le reazioni per questa conversazione in un'unica query
	//    e uniscile ai messaggi in Go.
	reactRows, err := db.c.Query(`
		SELECT r.id, r.messageId, r.emoji,
			u.id as reactorId, u.username as reactorUsername, u.photoUrl as reactorPhoto
		FROM reactions r
		JOIN users u ON r.userId = u.id
		WHERE r.messageId IN (SELECT id FROM messages WHERE conversationId = ?)
	`, conversationID)
	if err != nil {
		return conversation, fmt.Errorf("could not get reactions: %w", err)
	}
	defer reactRows.Close()

	for reactRows.Next() {
		var reaction Reaction
		var msgId string
		var reactorPhoto sql.NullString

		if err := reactRows.Scan(&reaction.ID, &msgId, &reaction.Emoji,
			&reaction.User.ID, &reaction.User.Username, &reactorPhoto); err != nil {
			return conversation, fmt.Errorf("could not scan reaction: %w", err)
		}
		reaction.User.PhotoURL = reactorPhoto.String

		// Aggiungi la reazione al messaggio corretto usando la mappa
		if msgPtr, ok := messageMap[msgId]; ok {
			msgPtr.Reactions = append(msgPtr.Reactions, reaction)
		}
	}

	conversation.Messages = messages

	if conversation.Members == nil {
		conversation.Members = []User{}
	}
	if conversation.Messages == nil {
		conversation.Messages = []Message{}
	}
	
	return conversation, nil
}


// GetConversationSummaries recupera la lista delle chat per un utente.
func (db *appdbimpl) GetConversationSummaries(userID string) ([]ConversationSummary, error) {
	var summaries []ConversationSummary

	// Questa query è la più complessa.
	// 1. Trova tutte le conversazioni (c) a cui l'utente (userID) partecipa.
	// 2. Per ogni conversazione, trova l'ultimo messaggio (lm).
	// 3. Calcola il nome e la foto:
	//    - Se è un gruppo (isGroup = 1), usa c.name e c.photoUrl.
	//    - Se è 1-a-1 (isGroup = 0), trova l'ALTRO utente (ou) e usa ou.username e ou.photoUrl.
	query := `
		-- 1. Definiamo l'ultimo messaggio per ogni conversazione
		WITH LatestMessages AS (
			SELECT
				conversationId,
				content,
				timestamp,
				ROW_NUMBER() OVER(PARTITION BY conversationId ORDER BY timestamp DESC) as rn
			FROM messages
			WHERE conversationId IN (SELECT conversationId FROM conversation_members WHERE userId = ?)
		),
		-- 2. Definiamo l'altro utente per le chat 1-a-1
		OtherUsers AS (
			SELECT 
				cm.conversationId, 
				u.username, 
				u.photoUrl 
			FROM conversation_members cm
			JOIN users u ON cm.userId = u.id
			WHERE cm.conversationId IN (SELECT conversationId FROM conversation_members WHERE userId = ?)
			  AND cm.userId != ?
		)
		-- 3. Uniamo tutto
		SELECT 
			c.id,
			c.isGroup,
			COALESCE(lm.content, '') AS latestMessageSnippet,
			COALESCE(lm.timestamp, '') AS latestMessageTimestamp,
			-- Se è un gruppo, usa il nome del gruppo, altrimenti il nome dell'altro utente
			CASE WHEN c.isGroup = 1 THEN c.name ELSE ou.username END AS conversationName,
			-- Se è un gruppo, usa la foto del gruppo, altrimenti la foto dell'altro utente
			CASE WHEN c.isGroup = 1 THEN c.photoUrl ELSE ou.photoUrl END AS conversationPhotoUrl
		FROM conversations c
		JOIN conversation_members cm_user ON c.id = cm_user.conversationId AND cm_user.userId = ?
		LEFT JOIN LatestMessages lm ON c.id = lm.conversationId AND lm.rn = 1
		LEFT JOIN OtherUsers ou ON c.id = ou.conversationId AND c.isGroup = 0
		ORDER BY latestMessageTimestamp DESC;
	`
	
	rows, err := db.c.Query(query, userID, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying conversation summaries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var summary ConversationSummary
		var isGroup bool
		var name, photo, snippet, timestamp sql.NullString

		err = rows.Scan(&summary.ID, &isGroup, &snippet, &timestamp, &name, &photo)
		if err != nil {
			return nil, fmt.Errorf("error scanning summary row: %w", err)
		}
		
		summary.Name = name.String
		summary.PhotoURL = photo.String
		summary.LatestMessageSnippet = snippet.String
		summary.LatestMessageTimestamp = timestamp.String
		
		summaries = append(summaries, summary)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating summary rows: %w", err)
	}

	return summaries, nil
}

// SendMessage crea un nuovo messaggio e lo restituisce.
func (db *appdbimpl) SendMessage(senderId string, convId string, content string, contentType string, replyToMsgId *string) (Message, error) {
	var message Message

	// 1. Iniziamo una transazione
	tx, err := db.c.Begin()
	if err != nil {
		return message, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback() // Annulla se qualcosa va storto

	// 2. [Controllo 403] L'utente è membro della conversazione?
	var isMember bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?)",
		convId, senderId).Scan(&isMember)
	if err != nil {
		return message, fmt.Errorf("error checking membership: %w", err)
	}
	if !isMember {
		return message, ErrForbidden // Errore 403
	}

	// 3. [Controllo 400] Se 'replyToMsgId' è fornito, esiste in QUESTA conversazione?
	if replyToMsgId != nil && *replyToMsgId != "" {
		var replyExists bool
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM messages WHERE id = ? AND conversationId = ?)",
			*replyToMsgId, convId).Scan(&replyExists)
		if err != nil {
			return message, fmt.Errorf("error checking reply message: %w", err)
		}
		if !replyExists {
			return message, ErrBadRequest // Errore 400
		}
	} else {
		// Assicurati che sia nil se la stringa è vuota, per il DB
		replyToMsgId = nil 
	}

	// 4. Crea il messaggio
	newMsgId := "msg-" + uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339Nano) // Formato ISO 8601

	_, err = tx.Exec(`INSERT INTO messages (id, conversationId, senderId, content, contentType, timestamp, replyToMsgId) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		newMsgId, convId, senderId, content, contentType, timestamp, replyToMsgId)
	if err != nil {
		return message, fmt.Errorf("error inserting message: %w", err)
	}

	// 5. Committa la transazione
	if err = tx.Commit(); err != nil {
		return message, fmt.Errorf("could not commit transaction: %w", err)
	}

	// 6. Recupera l'oggetto User del mittente (ci serve per la risposta)
	sender, err := db.GetUserByID(senderId)
	if err != nil {
		return message, fmt.Errorf("could not get sender details: %w", err)
	}

	// 7. Costruisci e restituisci l'oggetto Message completo (come da YAML)
	message = Message{
		ID:          newMsgId,
		Sender:      sender,
		Content:     content,
		ContentType: contentType,
		Timestamp:   timestamp,
		Reactions:   []Reaction{}, // Appena creato, non ha reazioni
		// 'status' e 'replyToMsgId' (struct) li omettiamo per ora
	}
	
	return message, nil
}

// DeleteMessage elimina un messaggio.
// Restituisce ErrForbidden se l'utente non è il mittente.
// Restituisce sql.ErrNoRows se il messaggio non esiste.
func (db *appdbimpl) DeleteMessage(requestingUserID string, messageID string) error {

	// 1. Controlliamo di chi è il messaggio
	var senderId string
	err := db.c.QueryRow("SELECT senderId FROM messages WHERE id = ?", messageID).Scan(&senderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Errore 404
			return sql.ErrNoRows 
		}
		// Altro errore
		return fmt.Errorf("error checking message sender: %w", err)
	}

	// 2. Controllo Autorizzazione (403)
	if senderId != requestingUserID {
		return ErrForbidden // Stesso errore ErrForbidden che abbiamo definito prima
	}

	// 3. L'utente è autorizzato. Elimina il messaggio.
	_, err = db.c.Exec("DELETE FROM messages WHERE id = ?", messageID)
	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}

	return nil // Successo
}

// ForwardMessage inoltra un messaggio esistente in una nuova conversazione.
func (db *appdbimpl) ForwardMessage(requestingUserID string, targetConvId string, originalMessageId string) (Message, error) {
    var originalMsg struct {
        Content     string
        ContentType string
    }
    var forwardedMessage Message

    // 1. Inizia transazione
    tx, err := db.c.Begin()
    if err != nil {
        return forwardedMessage, fmt.Errorf("could not begin transaction: %w", err)
    }
    defer tx.Rollback()

    // 2. [Controllo 403 Target] L'utente è membro della chat di destinazione?
    var isTargetMember bool
    err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?)",
        targetConvId, requestingUserID).Scan(&isTargetMember)
    if err != nil {
        return forwardedMessage, fmt.Errorf("error checking target membership: %w", err)
    }
    if !isTargetMember {
        // L'utente non è nella chat di destinazione
        return forwardedMessage, ErrForbidden 
    }

    // 3. [Controllo 404/403 Source] L'utente può vedere il messaggio originale?
    //    Recuperiamo il messaggio e verifichiamo che l'utente sia membro della chat *originale*.
    err = tx.QueryRow(`
        SELECT m.content, m.contentType
        FROM messages m
        JOIN conversation_members cm ON m.conversationId = cm.conversationId
        WHERE m.id = ? AND cm.userId = ?`,
        originalMessageId, requestingUserID).Scan(&originalMsg.Content, &originalMsg.ContentType)

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            // 404 (Messaggio non trovato o utente non membro della chat originale)
            return forwardedMessage, sql.ErrNoRows 
        }
        return forwardedMessage, fmt.Errorf("error getting original message: %w", err)
    }

    // 4. Crea il nuovo messaggio (l'inoltro) nella chat di destinazione
    newMsgId := "msg-" + uuid.New().String()
    timestamp := time.Now().UTC().Format(time.RFC3339Nano)

    _, err = tx.Exec(`
        INSERT INTO messages (id, conversationId, senderId, content, contentType, timestamp, forwardedFromMsgId)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
        newMsgId, targetConvId, requestingUserID, originalMsg.Content, originalMsg.ContentType, timestamp, originalMessageId)
    if err != nil {
        return forwardedMessage, fmt.Errorf("error inserting forwarded message: %w", err)
    }

    // 5. Committa
    if err = tx.Commit(); err != nil {
        return forwardedMessage, fmt.Errorf("could not commit transaction: %w", err)
    }

    // 6. Recupera i dettagli del mittente (per la risposta JSON)
    sender, err := db.GetUserByID(requestingUserID)
    if err != nil {
        return forwardedMessage, fmt.Errorf("could not get sender details: %w", err)
    }

    // 7. Costruisci e restituisci l'oggetto Message
    forwardedMessage = Message{
        ID:          newMsgId,
        Sender:      sender,
        Content:     originalMsg.Content,
        ContentType: originalMsg.ContentType,
        Timestamp:   timestamp,
        Reactions:   []Reaction{}, // Messaggio nuovo, no reazioni
    }

    return forwardedMessage, nil
}

// AddReaction aggiunge una reazione a un messaggio.
func (db *appdbimpl) AddReaction(requestingUserID string, messageID string, emoji string) (Reaction, error) {
    var reaction Reaction

    // 1. Controlla che l'emoji sia valida (esempio base)
    if len(emoji) == 0 || len(emoji) > 4 { // Emoji possono essere 4 byte
        return reaction, fmt.Errorf("emoji non valida: %w", ErrBadRequest)
    }

    // 2. Transazione
    tx, err := db.c.Begin()
    if err != nil {
        return reaction, fmt.Errorf("could not begin transaction: %w", err)
    }
    defer tx.Rollback()

    // 3. [Controllo 403] L'utente può vedere il messaggio?
    //    (è membro della conversazione del messaggio?)
    var isMember bool
    err = tx.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM conversation_members cm
            JOIN messages m ON cm.conversationId = m.conversationId
            WHERE m.id = ? AND cm.userId = ?
        )`, messageID, requestingUserID).Scan(&isMember)

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) { // Implicherebbe che il messaggio non esiste
            return reaction, sql.ErrNoRows // 404
        }
        return reaction, fmt.Errorf("error checking reaction permission: %w", err)
    }
    if !isMember {
        return reaction, ErrForbidden // 403
    }

    // 4. Inserisci o Sostituisci (UPSERT)
    // Cerchiamo prima se esiste già una reazione identica
    var existingId string
    err = tx.QueryRow(`SELECT id FROM reactions WHERE messageId = ? AND userId = ? AND emoji = ?`,
        messageID, requestingUserID, emoji).Scan(&existingId)

    if errors.Is(err, sql.ErrNoRows) {
        // Non esiste, crea
        reaction.ID = "react-" + uuid.New().String()
        _, err = tx.Exec(`INSERT INTO reactions (id, messageId, userId, emoji) VALUES (?, ?, ?, ?)`,
            reaction.ID, messageID, requestingUserID, emoji)
    } else if err == nil {
        // Esiste già, usa l'ID esistente
        reaction.ID = existingId
    } else {
        // Errore
        return reaction, fmt.Errorf("error checking existing reaction: %w", err)
    }

    if err != nil {
        return reaction, fmt.Errorf("error upserting reaction: %w", err)
    }

    // 5. Committa
    if err = tx.Commit(); err != nil {
        return reaction, fmt.Errorf("could not commit transaction: %w", err)
    }

    // 6. Costruisci la risposta
    user, err := db.GetUserByID(requestingUserID) 
    if err != nil {
        return reaction, fmt.Errorf("could not get reactor user details: %w", err)
    }

    reaction.Emoji = emoji
    reaction.User = user

    return reaction, nil
}

// RemoveReaction elimina una reazione.
func (db *appdbimpl) RemoveReaction(requestingUserID string, reactionID string, messageID string) error {
    // Esegui la cancellazione solo se l'ID reazione, l'ID messaggio
    // e l'ID utente (proprietario) corrispondono.
    // Questo previene che un utente cancelli la reazione di un altro (403).
    res, err := db.c.Exec(`
        DELETE FROM reactions 
        WHERE id = ? AND messageId = ? AND userId = ?`,
        reactionID, messageID, requestingUserID)

    if err != nil {
        return fmt.Errorf("error deleting reaction: %w", err)
    }

    // Controlla se qualche riga è stata effettivamente cancellata
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("error checking affected rows: %w", err)
    }

    if rowsAffected == 0 {
        // 404 (non trovato) o 403 (non è tuo)
        return sql.ErrNoRows 
    }

    return nil // Successo
}

// CreateGroup crea una nuova conversazione di gruppo.
func (db *appdbimpl) CreateGroup(requestingUserID string, groupName string, memberIds []string) (string, error) {
    tx, err := db.c.Begin()
    if err != nil {
        return "", fmt.Errorf("could not begin transaction: %w", err)
    }
    defer tx.Rollback()

    // 1. Crea la conversazione (con isGroup = 1)
    newConvID := "conv-" + uuid.New().String()
    _, err = tx.Exec("INSERT INTO conversations (id, name, isGroup) VALUES (?, ?, 1)", newConvID, groupName)
    if err != nil {
        return "", fmt.Errorf("could not create group conversation: %w", err)
    }

    // 2. Aggiungi il creatore al gruppo
    _, err = tx.Exec("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)", newConvID, requestingUserID)
    if err != nil {
        return "", fmt.Errorf("could not add creator to group: %w", err)
    }

    // 3. Aggiungi tutti gli altri membri
    stmt, err := tx.Prepare("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)")
    if err != nil {
        return "", fmt.Errorf("could not prepare member insert: %w", err)
    }
    defer stmt.Close()

    for _, memberId := range memberIds {
        if _, err = stmt.Exec(newConvID, memberId); err != nil {
            // Se l'ID utente non esiste, questo fallirà (FOREIGN KEY constraint)
            return "", fmt.Errorf("could not add member %s: %w", memberId, err)
        }
    }

    // 4. Committa
    if err = tx.Commit(); err != nil {
        return "", fmt.Errorf("could not commit transaction: %w", err)
    }

    return newConvID, nil
}

// checkGroupAccess verifica se un utente è membro di un gruppo.
// Restituisce ErrForbidden se non è membro, ErrBadRequest se non è un gruppo.
func (db *appdbimpl) checkGroupAccess(tx *sql.Tx, requestingUserID string, convId string) error {
    var isGroup bool
    var isMember bool

    // Usiamo COALESCE per gestire i NULL (in caso di subquery vuote)
    query := `
        SELECT
            (SELECT isGroup FROM conversations WHERE id = ?) AS isGroup,
            EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?) AS isMember`
    
    // Scegliamo se usare la transazione (tx) o la connessione (db.c)
    var row *sql.Row
    if tx != nil {
        row = tx.QueryRow(query, convId, convId, requestingUserID)
    } else {
        row = db.c.QueryRow(query, convId, convId, requestingUserID)
    }

    if err := row.Scan(&isGroup, &isMember); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return sql.ErrNoRows // 404
        }
        return err // 500
    }

    if !isMember {
        return ErrForbidden // 403
    }
    if !isGroup {
        return ErrBadRequest // 400 (o 403 a seconda della logica)
    }
    return nil // Accesso consentito
}


// SetGroupName aggiorna il nome di un gruppo.
func (db *appdbimpl) SetGroupName(requestingUserID string, convId string, newName string) error {
    // 1. Controlla i permessi
    if err := db.checkGroupAccess(nil, requestingUserID, convId); err != nil {
        return err // Restituisce 403, 404, o 400
    }

    // 2. Aggiorna il nome
    _, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ?", newName, convId)
    if err != nil {
        return fmt.Errorf("error updating group name: %w", err)
    }
    return nil
}

// SetGroupPhoto aggiorna la foto di un gruppo.
func (db *appdbimpl) SetGroupPhoto(requestingUserID string, convId string, newPhotoURL string) error {
    // 1. Controlla i permessi
    if err := db.checkGroupAccess(nil, requestingUserID, convId); err != nil {
        return err // Restituisce 403, 404, o 400
    }

    // 2. Aggiorna la foto
    _, err := db.c.Exec("UPDATE conversations SET photoUrl = ? WHERE id = ?", newPhotoURL, convId)
    if err != nil {
        return fmt.Errorf("error updating group photo: %w", err)
    }
    return nil
}

// AddGroupMember aggiunge un utente a un gruppo.
func (db *appdbimpl) AddGroupMember(requestingUserID string, convId string, targetUserID string) error {
    tx, err := db.c.Begin()
    if err != nil {
        return fmt.Errorf("could not begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    // 1. Controlla i permessi
    if err := db.checkGroupAccess(tx, requestingUserID, convId); err != nil {
        return err // Restituisce 403, 404, o 400
    }

    // 2. Controlla che l'utente target esista (per 404)
    exists, err := db.CheckUserExists(targetUserID)
    if err != nil {
        return err
    }
    if !exists {
        return sql.ErrNoRows // 404
    }

    // 3. Inserisci il nuovo membro
    _, err = tx.Exec("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)", convId, targetUserID)
    if err != nil {
        // Controlla se l'errore è "UNIQUE constraint failed"
        var sqliteErr sqlite3.Error
        if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
            return ErrAlreadyMember // 409
        }
        return fmt.Errorf("error adding member: %w", err)
    }

    return tx.Commit()
}

// LeaveGroup rimuove l'utente autenticato da un gruppo.
func (db *appdbimpl) LeaveGroup(requestingUserID string, convId string) error {
    tx, err := db.c.Begin()
    if err != nil {
        return fmt.Errorf("could not begin transaction: %w", err)
    }
    defer tx.Rollback()

    // 1. Controlla i permessi (verifica che sia membro e che sia un gruppo)
    if err := db.checkGroupAccess(tx, requestingUserID, convId); err != nil {
        return err
    }

    // 2. Rimuovi il membro
    _, err = tx.Exec("DELETE FROM conversation_members WHERE conversationId = ? AND userId = ?", convId, requestingUserID)
    if err != nil {
        return fmt.Errorf("error leaving group: %w", err)
    }
    
    // (Logica opzionale: se il gruppo è vuoto, cancellarlo? Per ora no)
    
    return tx.Commit()
}