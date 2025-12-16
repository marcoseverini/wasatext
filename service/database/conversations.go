package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (db *appdbimpl) StartConversation(requestingUserID string, targetUserID string) (string, error) {
	// Trova una chat 1-1 esistente o ne crea una nuova.
	// Restituisce l'ID della conversazione.

	// Cerca una chat 1-1 (non di gruppo) esistente tra questi due utenti.
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
		return existingConvID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("error finding existing conversation: %w", err)
	}

	// Se la conversazione non è stata trovata ne crea una nuova.
	tx, err := db.c.Begin()
	if err != nil {
		return "", fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Crea la nuova conversazione
	newConvID := "conv-" + uuid.New().String()
	_, err = tx.Exec("INSERT INTO conversations (id, isGroup) VALUES (?, 0)", newConvID)
	if err != nil {
		return "", fmt.Errorf("could not create conversation: %w", err)
	}

	// Aggiunge i membri
	_, err = tx.Exec("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)", newConvID, requestingUserID)
	if err != nil {
		return "", fmt.Errorf("could not add requesting user: %w", err)
	}
	_, err = tx.Exec("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)", newConvID, targetUserID)
	if err != nil {
		return "", fmt.Errorf("could not add target user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("could not commit transaction: %w", err)
	}

	return newConvID, nil
}

func (db *appdbimpl) GetConversationDetails(conversationID string, requestingUserID string) (Conversation, error) {
	var conversation Conversation

	// 1. Verifica Membro
	var isMember bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?)", conversationID, requestingUserID).Scan(&isMember)
	if err != nil || !isMember {
		return conversation, fmt.Errorf("user not member or conversation not found")
	}

	// 2. Aggiorna stato lettura
	_, _ = db.c.Exec(`UPDATE messages SET status = 'read' WHERE conversationId = ? AND senderId != ? AND status != 'read'`, conversationID, requestingUserID)

	// 3. Dettagli base conversazione
	var nullableName sql.NullString
	var nullablePhoto sql.NullString
	err = db.c.QueryRow("SELECT id, name, photoUrl, isGroup FROM conversations WHERE id = ?", conversationID).
		Scan(&conversation.ID, &nullableName, &nullablePhoto, &conversation.IsGroup)
	if err != nil {
		return conversation, fmt.Errorf("could not get conversation details: %w", err)
	}
	conversation.Name = nullableName.String
	conversation.PhotoURL = nullablePhoto.String

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

	// 4. Recupera Membri
	rows, err := db.c.Query(`SELECT u.id, u.username, u.photoUrl FROM users u JOIN conversation_members cm ON u.id = cm.userId WHERE cm.conversationId = ?`, conversationID)
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
	if err = rows.Err(); err != nil {
		return conversation, err
	}
	conversation.Members = members

	// 5. Recupera Messaggi
	msgRows, err := db.c.Query(`
        SELECT m.id, COALESCE(m.text, ''), COALESCE(m.photoUrl, ''), m.timestamp, m.replyToMsgId, m.status,
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
	messageMap := make(map[string]int)

	for msgRows.Next() {
		var msg Message
		var senderPhoto sql.NullString
		var replyTo sql.NullString

		if err := msgRows.Scan(&msg.ID, &msg.Text, &msg.PhotoURL, &msg.Timestamp, &replyTo, &msg.Status,
			&msg.Sender.ID, &msg.Sender.Username, &senderPhoto); err != nil {
			return conversation, fmt.Errorf("could not scan message: %w", err)
		}
		msg.Sender.PhotoURL = senderPhoto.String

		if replyTo.Valid {
			val := replyTo.String
			msg.ReplyToMsgId = &val
		}

		msg.Reactions = []Reaction{}
		messages = append(messages, msg)
		messageMap[msg.ID] = len(messages) - 1
	}
	if err = msgRows.Err(); err != nil {
		return conversation, err
	}

	// 6. Recupera Reazioni
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

		if idx, ok := messageMap[msgId]; ok {
			messages[idx].Reactions = append(messages[idx].Reactions, reaction)
		}
	}
	if err = reactRows.Err(); err != nil {
		return conversation, err
	}

	conversation.Messages = messages
	if conversation.Members == nil {
		conversation.Members = []User{}
	}
	if conversation.Messages == nil {
		conversation.Messages = []Message{}
	}

	// --- AGGIUNTO: Popoliamo i campi Snippet e Timestamp per coerenza con l'API ---
	if len(messages) > 0 {
		lastMsg := messages[0]
		conversation.LatestMessageTimestamp = lastMsg.Timestamp
		if lastMsg.Text != "" {
			conversation.LatestMessageSnippet = lastMsg.Text
		} else if lastMsg.PhotoURL != "" {
			conversation.LatestMessageSnippet = "📷 [Foto]"
		}
	}
	// -------------------------------------------------------------------------------

	return conversation, nil
}

// Recupera la lista delle chat per un utente.
func (db *appdbimpl) GetConversationSummaries(userID string) ([]ConversationSummary, error) {

	var summaries []ConversationSummary

	// QUERY AGGIORNATA: Genera l'anteprima (snippet) usando text o photoUrl
	// Se c'è del testo, mostriamo quello.
	// Se c'è solo una foto, mostriamo "[Foto]".
	// Se non c'è nulla, stringa vuota.
	query := `
		WITH LatestMessages AS (
			SELECT
				conversationId,
				CASE 
					WHEN text IS NOT NULL AND text != '' THEN text
					WHEN photoUrl IS NOT NULL AND photoUrl != '' THEN '📷 [Foto]'
					ELSE ''
				END as content,
				timestamp,
				ROW_NUMBER() OVER(PARTITION BY conversationId ORDER BY timestamp DESC) as rn
			FROM messages
			WHERE conversationId IN (SELECT conversationId FROM conversation_members WHERE userId = ?)
		),
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
		SELECT 
			c.id,
			c.isGroup,
			COALESCE(lm.content, '') AS latestMessageSnippet,
			COALESCE(lm.timestamp, '') AS latestMessageTimestamp,
			CASE WHEN c.isGroup = 1 THEN c.name ELSE ou.username END AS conversationName,
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
