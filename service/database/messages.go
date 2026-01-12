package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Crea un nuovo messaggio con testo E/O foto.
func (db *appdbimpl) SendMessage(senderId string, convId string, text string, photoUrl string, replyToMsgId *string) (Message, error) {

	var message Message

	tx, err := db.c.Begin()
	if err != nil {
		return message, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Check Membership
	var isMember bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?)",
		convId, senderId).Scan(&isMember)
	if err != nil {
		return message, fmt.Errorf("error checking membership: %w", err)
	}
	if !isMember {
		return message, ErrForbidden
	}

	// Check Reply
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

	// Insert Message
	newMsgId := "msg-" + uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	_, err = tx.Exec(`INSERT INTO messages (id, conversationId, senderId, text, photoUrl, timestamp, replyToMsgId, status) 
		VALUES (?, ?, ?, ?, ?, ?, ?, 'sent')`,
		newMsgId, convId, senderId, text, photoUrl, timestamp, replyToMsgId)
	if err != nil {
		return message, fmt.Errorf("error inserting message: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return message, fmt.Errorf("could not commit transaction: %w", err)
	}

	sender, err := db.GetUserByID(senderId)
	if err != nil {
		return message, fmt.Errorf("could not get sender details: %w", err)
	}

	message = Message{
		ID:           newMsgId,
		Sender:       sender,
		Text:         text,
		PhotoURL:     photoUrl,
		Timestamp:    timestamp,
		Status:       "sent",
		ReplyToMsgId: replyToMsgId,
		Reactions:    []Reaction{},
	}

	return message, nil
}

func (db *appdbimpl) DeleteMessage(requestingUserID string, messageID string) error {
	var senderId string
	err := db.c.QueryRow("SELECT senderId FROM messages WHERE id = ?", messageID).Scan(&senderId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return fmt.Errorf("error checking message sender: %w", err)
	}
	if senderId != requestingUserID {
		return ErrForbidden
	}
	_, err = db.c.Exec("DELETE FROM messages WHERE id = ?", messageID)
	if err != nil {
		return fmt.Errorf("error deleting message: %w", err)
	}
	return nil
}

func (db *appdbimpl) ForwardMessage(requestingUserID string, targetConvId string, originalMessageId string) (Message, error) {

	var originalMsg struct {
		Text     string
		PhotoURL string
	}
	var forwardedMessage Message

	tx, err := db.c.Begin()
	if err != nil {
		return forwardedMessage, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var isTargetMember bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?)",
		targetConvId, requestingUserID).Scan(&isTargetMember)
	if err != nil {
		return forwardedMessage, fmt.Errorf("error checking target membership: %w", err)
	}
	if !isTargetMember {
		return forwardedMessage, ErrForbidden
	}

	// Fetch original message (Text + Photo)
	err = tx.QueryRow(`
        SELECT COALESCE(m.text, ''), COALESCE(m.photoUrl, '')
        FROM messages m
        JOIN conversation_members cm ON m.conversationId = cm.conversationId
        WHERE m.id = ? AND cm.userId = ?`,
		originalMessageId, requestingUserID).Scan(&originalMsg.Text, &originalMsg.PhotoURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return forwardedMessage, sql.ErrNoRows
		}
		return forwardedMessage, fmt.Errorf("error getting original message: %w", err)
	}

	newMsgId := "msg-" + uuid.New().String()
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)

	_, err = tx.Exec(`
        INSERT INTO messages (id, conversationId, senderId, text, photoUrl, timestamp, forwardedFromMsgId)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		newMsgId, targetConvId, requestingUserID, originalMsg.Text, originalMsg.PhotoURL, timestamp, originalMessageId)
	if err != nil {
		return forwardedMessage, fmt.Errorf("error inserting forwarded message: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return forwardedMessage, fmt.Errorf("could not commit transaction: %w", err)
	}

	sender, err := db.GetUserByID(requestingUserID)
	if err != nil {
		return forwardedMessage, fmt.Errorf("could not get sender details: %w", err)
	}

	forwardedMessage = Message{
		ID:        newMsgId,
		Sender:    sender,
		Text:      originalMsg.Text,
		PhotoURL:  originalMsg.PhotoURL,
		Timestamp: timestamp,
		Reactions: []Reaction{},
	}

	return forwardedMessage, nil
}

func (db *appdbimpl) AddReaction(requestingUserID string, messageID string, emoji string) (Reaction, error) {
	var reaction Reaction
	if len(emoji) == 0 || len(emoji) > 8 {
		return reaction, fmt.Errorf("emoji non valida: %w", ErrBadRequest)
	}
	tx, err := db.c.Begin()
	if err != nil {
		return reaction, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var isMember bool
	err = tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM conversation_members cm JOIN messages m ON cm.conversationId = m.conversationId WHERE m.id = ? AND cm.userId = ?)`, messageID, requestingUserID).Scan(&isMember)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return reaction, sql.ErrNoRows
		}
		return reaction, fmt.Errorf("error checking reaction permission: %w", err)
	}
	if !isMember {
		return reaction, ErrForbidden
	}

	var existingId string
	err = tx.QueryRow(`SELECT id FROM reactions WHERE messageId = ? AND userId = ?`, messageID, requestingUserID).Scan(&existingId)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		reaction.ID = "react-" + uuid.New().String()
		_, err = tx.Exec(`INSERT INTO reactions (id, messageId, userId, emoji) VALUES (?, ?, ?, ?)`, reaction.ID, messageID, requestingUserID, emoji)
		if err != nil {
			return reaction, fmt.Errorf("error inserting reaction: %w", err)
		}
	case err == nil:
		reaction.ID = existingId
		_, err = tx.Exec(`UPDATE reactions SET emoji = ? WHERE id = ?`, emoji, existingId)
		if err != nil {
			return reaction, fmt.Errorf("error updating reaction: %w", err)
		}
	default:
		return reaction, fmt.Errorf("error checking existing reaction: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return reaction, fmt.Errorf("could not commit transaction: %w", err)
	}

	user, err := db.GetUserByID(requestingUserID)
	if err != nil {
		return reaction, fmt.Errorf("could not get reactor user details: %w", err)
	}
	reaction.Emoji = emoji
	reaction.User = user
	return reaction, nil
}

func (db *appdbimpl) RemoveReaction(requestingUserID string, reactionID string, messageID string) error {
	var ownerId string
	err := db.c.QueryRow("SELECT userId FROM reactions WHERE id = ? AND messageId = ?", reactionID, messageID).Scan(&ownerId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return fmt.Errorf("error checking reaction owner: %w", err)
	}
	if ownerId != requestingUserID {
		return ErrForbidden
	}
	_, err = db.c.Exec("DELETE FROM reactions WHERE id = ?", reactionID)
	if err != nil {
		return fmt.Errorf("error deleting reaction: %w", err)
	}
	return nil
}
