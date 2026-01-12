package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

// Crea una nuova conversazione di gruppo.
func (db *appdbimpl) CreateGroup(requestingUserID string, groupName string, memberIds []string) (string, error) {
	tx, err := db.c.Begin()
	if err != nil {
		return "", fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	// Annulla se qualcosa va storto

	// Crea la conversazione (con isGroup = 1)
	newConvID := "conv-" + uuid.New().String()
	_, err = tx.Exec("INSERT INTO conversations (id, name, isGroup) VALUES (?, ?, 1)", newConvID, groupName)
	if err != nil {
		return "", fmt.Errorf("could not create group conversation: %w", err)
	}

	// Aggiunge il creatore al gruppo
	_, err = tx.Exec("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)", newConvID, requestingUserID)
	if err != nil {
		return "", fmt.Errorf("could not add creator to group: %w", err)
	}

	// Aggiunge tutti gli altri membri
	stmt, err := tx.Prepare("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)")
	if err != nil {
		return "", fmt.Errorf("could not prepare member insert: %w", err)
	}
	defer stmt.Close()

	for _, memberId := range memberIds {
		if _, err = stmt.Exec(newConvID, memberId); err != nil {
			// Se l'ID utente non esiste, questo fallirà
			return "", fmt.Errorf("could not add member %s: %w", memberId, err)
		}
	}

	// 4. Committa
	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("could not commit transaction: %w", err)
	}

	return newConvID, nil
}

// Verifica se un utente è membro di un gruppo.
// Restituisce ErrForbidden se non è membro, ErrBadRequest se non è un gruppo.
func (db *appdbimpl) checkGroupAccess(tx *sql.Tx, requestingUserID string, convId string) error {

	var isGroup bool
	var isMember bool

	query := `
        SELECT
            (SELECT isGroup FROM conversations WHERE id = ?) AS isGroup,
            EXISTS(SELECT 1 FROM conversation_members WHERE conversationId = ? AND userId = ?) AS isMember`

	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, convId, convId, requestingUserID)
	} else {
		row = db.c.QueryRow(query, convId, convId, requestingUserID)
	}

	if err := row.Scan(&isGroup, &isMember); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return sql.ErrNoRows
		default:
			return err
		}
	}

	if !isMember {
		return ErrForbidden
	}
	if !isGroup {
		return ErrBadRequest
	}
	return nil // Accesso consentito
}

// Aggiorna il nome di un gruppo.
func (db *appdbimpl) SetGroupName(requestingUserID string, convId string, newName string) error {

	// Controlla i permessi
	if err := db.checkGroupAccess(nil, requestingUserID, convId); err != nil {
		return err
	}

	// Aggiorna il nome
	_, err := db.c.Exec("UPDATE conversations SET name = ? WHERE id = ?", newName, convId)
	if err != nil {
		return fmt.Errorf("error updating group name: %w", err)
	}
	return nil
}

// Aggiorna la foto di un gruppo.
func (db *appdbimpl) SetGroupPhoto(requestingUserID string, convId string, newPhotoURL string) error {

	// Controlla i permessi
	if err := db.checkGroupAccess(nil, requestingUserID, convId); err != nil {
		return err
	}

	// Aggiorna la foto
	_, err := db.c.Exec("UPDATE conversations SET photoUrl = ? WHERE id = ?", newPhotoURL, convId)
	if err != nil {
		return fmt.Errorf("error updating group photo: %w", err)
	}
	return nil
}

// Aggiunge un utente a un gruppo.
func (db *appdbimpl) AddGroupMember(requestingUserID string, convId string, targetUserID string) error {

	tx, err := db.c.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	// Annulla se qualcosa va storto

	// Controlla i permessi
	if err := db.checkGroupAccess(tx, requestingUserID, convId); err != nil {
		return err
	}

	// Controlla che l'utente target esista
	exists, err := db.CheckUserExists(targetUserID)
	if err != nil {
		return err
	}
	if !exists {
		return sql.ErrNoRows
	}

	// Inserisce il nuovo membro
	_, err = tx.Exec("INSERT INTO conversation_members (conversationId, userId) VALUES (?, ?)", convId, targetUserID)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
				return ErrAlreadyMember
			}
		}
		return fmt.Errorf("error adding member: %w", err)
	}

	return tx.Commit()
}

// Rimuove l'utente autenticato da un gruppo.
func (db *appdbimpl) LeaveGroup(requestingUserID string, convId string) error {

	tx, err := db.c.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	// Annulla se qualcosa va storto

	// Controlla i permessi (verifica che sia membro e che sia un gruppo)
	if err := db.checkGroupAccess(tx, requestingUserID, convId); err != nil {
		return err
	}

	// Rimuove il membro
	_, err = tx.Exec("DELETE FROM conversation_members WHERE conversationId = ? AND userId = ?", convId, requestingUserID)
	if err != nil {
		return fmt.Errorf("error leaving group: %w", err)
	}

	// (opzionale: se il gruppo è vuoto, cancellarlo? Per ora no)

	return tx.Commit()
}
