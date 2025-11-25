package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Crea un nuovo messaggio e lo restituisce.
func (db *appdbimpl) SendMessage(senderId string, convId string, content string, contentType string, replyToMsgId *string) (Message, error) {

	var message Message // components/schemas/Message

	// Inizia una transazione
	tx, err := db.c.Begin()
	if err != nil {
		return message, fmt.Errorf("could not begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()
	// Annulla se qualcosa va storto

	// Controlla che l'utente sia membro della conversazione
	var isMember bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?)",
		convId, senderId).Scan(&isMember)
	if err != nil {
		return message, fmt.Errorf("error checking membership: %w", err)
	}
	if !isMember {
		return message, ErrForbidden
	}

	// Controlla, se 'replyToMsgId' è fornito, se esiste in questa conversazione
	if replyToMsgId != nil && *replyToMsgId != "" {
		var replyExists bool
		err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM messages WHERE id = ? AND conversationId = ?)",
			*replyToMsgId, convId).Scan(&replyExists)
		if err != nil {
			return message, fmt.Errorf("error checking reply message: %w", err)
		}
		if !replyExists {
			return message, ErrBadRequest
		}
	} else {
		replyToMsgId = nil
	}

	// Crea il messaggio
	newMsgId := "msg-" + uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339Nano) // Formato ISO 8601

	_, err = tx.Exec(`INSERT INTO messages (id, conversationId, senderId, content, contentType, timestamp, replyToMsgId, status) 
		VALUES (?, ?, ?, ?, ?, ?, ?, 'sent')`,
		newMsgId, convId, senderId, content, contentType, timestamp, replyToMsgId)
	if err != nil {
		return message, fmt.Errorf("error inserting message: %w", err)
	}

	// Committa la transazione
	if err = tx.Commit(); err != nil {
		return message, fmt.Errorf("could not commit transaction: %w", err)
	}

	// Recupera l'oggetto User del mittente
	sender, err := db.GetUserByID(senderId)
	if err != nil {
		return message, fmt.Errorf("could not get sender details: %w", err)
	}

	// Costruisce e restituisce l'oggetto Message completo
	message = Message{ // components/schemas/Message
		ID:           newMsgId,
		Sender:       sender,
		Content:      content,
		ContentType:  contentType,
		Timestamp:    timestamp,
		Status:       "sent",       //
		ReplyToMsgId: replyToMsgId, // (Se non c'era già)
		Reactions:    []Reaction{},
	}

	return message, nil
}

// Elimina un messaggio.
func (db *appdbimpl) DeleteMessage(requestingUserID string, messageID string) error {

	// Controlla di chi è il messaggio
	var senderId string
	err := db.c.QueryRow("SELECT senderId FROM messages WHERE id = ?", messageID).Scan(&senderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return fmt.Errorf("error checking message sender: %w", err)
	}

	// Controllo Autorizzazione
	if senderId != requestingUserID {
		return ErrForbidden
	}

	// L'utente è autorizzato. Elimina il messaggio.
	_, err = db.c.Exec("DELETE FROM messages WHERE id = ?", messageID)
	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}

	return nil // Successo
}

// Inoltra un messaggio esistente in una nuova conversazione.
func (db *appdbimpl) ForwardMessage(requestingUserID string, targetConvId string, originalMessageId string) (Message, error) {

	var originalMsg struct {
		Content     string
		ContentType string
	}
	var forwardedMessage Message // components/schemas/Message

	// Inizia transazione
	tx, err := db.c.Begin()
	if err != nil {
		return forwardedMessage, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}() // Annulla se qualcosa va storto

	// Controlla se l'utente è membro della chat di destinazione
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

	// Recupera il messaggio e verifica che l'utente sia membro della chat originale.
	err = tx.QueryRow(`
        SELECT m.content, m.contentType
        FROM messages m
        JOIN conversation_members cm ON m.conversationId = cm.conversationId
        WHERE m.id = ? AND cm.userId = ?`,
		originalMessageId, requestingUserID).Scan(&originalMsg.Content, &originalMsg.ContentType)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return forwardedMessage, sql.ErrNoRows
		}
		return forwardedMessage, fmt.Errorf("error getting original message: %w", err)
	}

	// Crea il nuovo messaggio (l'inoltro) nella chat di destinazione
	newMsgId := "msg-" + uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	_, err = tx.Exec(`
        INSERT INTO messages (id, conversationId, senderId, content, contentType, timestamp, forwardedFromMsgId)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		newMsgId, targetConvId, requestingUserID, originalMsg.Content, originalMsg.ContentType, timestamp, originalMessageId)
	if err != nil {
		return forwardedMessage, fmt.Errorf("error inserting forwarded message: %w", err)
	}

	// Committa
	if err = tx.Commit(); err != nil {
		return forwardedMessage, fmt.Errorf("could not commit transaction: %w", err)
	}

	// Recupera i dettagli del mittente
	sender, err := db.GetUserByID(requestingUserID)
	if err != nil {
		return forwardedMessage, fmt.Errorf("could not get sender details: %w", err)
	}

	// Costruisce e restituisci l'oggetto Message
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

// Aggiunge o Aggiorna una reazione a un messaggio (Max 1 per utente)
func (db *appdbimpl) AddReaction(requestingUserID string, messageID string, emoji string) (Reaction, error) {

	var reaction Reaction

	// Validazione lunghezza (quella che abbiamo corretto a 8)
	if len(emoji) == 0 || len(emoji) > 8 {
		return reaction, fmt.Errorf("emoji non valida: %w", ErrBadRequest)
	}

	// Transazione
	tx, err := db.c.Begin()
	if err != nil {
		return reaction, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. Check Membership (Invariato)
	var isMember bool
	err = tx.QueryRow(`
        SELECT EXISTS (
            SELECT 1 FROM conversation_members cm
            JOIN messages m ON cm.conversationId = m.conversationId
            WHERE m.id = ? AND cm.userId = ?
        )`, messageID, requestingUserID).Scan(&isMember)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return reaction, sql.ErrNoRows
		}
		return reaction, fmt.Errorf("error checking reaction permission: %w", err)
	}
	if !isMember {
		return reaction, ErrForbidden
	}

	// 2. NUOVA LOGICA: "Upsert" (Update or Insert)
	// Cerchiamo se l'utente ha GIÀ una reazione (di qualsiasi tipo) su questo messaggio
	var existingId string
	err = tx.QueryRow(`SELECT id FROM reactions WHERE messageId = ? AND userId = ?`,
		messageID, requestingUserID).Scan(&existingId)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		// CASO A: Non esiste nessuna reazione -> CREIAMO NUOVA
		reaction.ID = "react-" + uuid.New().String()
		_, err = tx.Exec(`INSERT INTO reactions (id, messageId, userId, emoji) VALUES (?, ?, ?, ?)`,
			reaction.ID, messageID, requestingUserID, emoji)
		if err != nil {
			return reaction, fmt.Errorf("error inserting reaction: %w", err)
		}

	case err == nil:
		// CASO B: Esiste già una reazione -> AGGIORNIAMO L'EMOJI
		reaction.ID = existingId
		_, err = tx.Exec(`UPDATE reactions SET emoji = ? WHERE id = ?`, emoji, existingId)
		if err != nil {
			return reaction, fmt.Errorf("error updating reaction: %w", err)
		}

	default:
		// CASO C: Errore generico database
		return reaction, fmt.Errorf("error checking existing reaction: %w", err)
	}

	// Committa
	if err = tx.Commit(); err != nil {
		return reaction, fmt.Errorf("could not commit transaction: %w", err)
	}

	// Costruisce la risposta
	user, err := db.GetUserByID(requestingUserID)
	if err != nil {
		return reaction, fmt.Errorf("could not get reactor user details: %w", err)
	}

	reaction.Emoji = emoji
	reaction.User = user

	return reaction, nil
}

// Elimina una reazione.
func (db *appdbimpl) RemoveReaction(requestingUserID string, reactionID string, messageID string) error {

	// Controlla prima l'esistenza e la proprietà
	var ownerId string
	err := db.c.QueryRow("SELECT userId FROM reactions WHERE id = ? AND messageId = ?",
		reactionID, messageID).Scan(&ownerId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// La reazione (o il messaggio) non esiste
			return sql.ErrNoRows
		}
		return fmt.Errorf("error checking reaction owner: %w", err)
	}

	// Controllo Autorizzazione
	if ownerId != requestingUserID {
		// La reazione esiste, ma non è tua
		return ErrForbidden
	}

	// L'utente è autorizzato. Ora elimina il messaggio.
	_, err = db.c.Exec("DELETE FROM reactions WHERE id = ?", reactionID)
	if err != nil {
		return fmt.Errorf("error deleting reaction: %w", err) // 500
	}

	return nil
}
