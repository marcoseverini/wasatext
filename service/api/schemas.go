package api

import (
	"errors"  // Libreria per gestire gli errori
	"net/url" // Libreria per gestire gli URL
	"regexp"  // Libreria per le espressioni regolari
)

var (
	uuidRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
)

// Tipi Base e loro validazioni

type UserID string // components/schemas/UserID

func (id UserID) Validate() error {
	// Prende una stringa e verifica se è un UserID valido

	if !uuidRegex.MatchString(string(id)) {
		return errors.New("Formato UserID non valido. Deve essere un UUID.")
	}
	return nil
}

type InternalID string // components/schemas/InternalID
// Usato per ID interni come MessageID, GroupID, ecc.

func (id InternalID) Validate() error {
	// Prende una stringa e verifica se è un InternalID valido

	if len(id) < 1 || len(id) > 50 {
		return errors.New("ID interno non valido (deve essere tra 1 e 50 caratteri)")
	}
	return nil
}

type Username string // components/schemas/Username

func (u Username) Validate() error {
	// Prende una stringa e verifica se è un Username valido

	if len(u) < 3 || len(u) > 16 {
		return errors.New("Nome utente non valido (deve essere tra 3 e 16 caratteri)")
	}
	return nil
}

type PhotoURL string // components/schemas/PhotoURL

func (p PhotoURL) Validate() error {
	// Prende una stringa e verifica se è un PhotoURL valido

	if len(p) < 1 || len(p) > 2048 {
		return errors.New("URL foto non valido (lunghezza max 2048)")
	}
	if _, err := url.ParseRequestURI(string(p)); err != nil {
		return errors.New("URL foto non valido (formato URI non corretto)")
	}
	return nil
}

type Emoji string // components/schemas/Emoji

func (e Emoji) Validate() error {
	// Prende una stringa e verifica se è un Emoji valido

	if len(e) < 1 || len(e) > 4 {
		return errors.New("Emoji non valido (deve essere tra 1 e 4 caratteri)")
	}
	return nil
}

type Timestamp string // components/schemas/Timestamp
// La validazione è 'format: date-time', la gestisce il JSON decoder

type GroupName string // components/schemas/GroupName

func (g GroupName) Validate() error {
	// Prende una stringa e verifica se è un GroupName valido

	if len(g) < 1 || len(g) > 50 {
		return errors.New("Nome gruppo non valido (deve essere tra 1 e 50 caratteri)")
	}
	return nil
}

type MessageContent string // components/schemas/MessageContent

func (m MessageContent) Validate() error {
	// Prende una stringa e verifica se è un MessageContent valido

	if len(m) < 1 || len(m) > 4000 {
		return errors.New("Contenuto messaggio non valido (deve essere tra 1 e 4000 caratteri)")
	}
	return nil
}

type SearchQuery string

func (s SearchQuery) Validate() error {
	// Prende una stringa e verifica se è una query di ricerca valida
	// (come da YAML: min 1, max 16)
	if len(s) < 1 || len(s) > 16 {
		return errors.New("il termine di ricerca deve essere tra 1 e 16 caratteri")
	}
	return nil
}

// Schemi di richiesta/risposta API

type ErrorResponse struct { // components/schemas/ErrorResponse
	Message string `json:"message"`
}

type LoginRequest struct { // components/schemas/LoginRequest
	Username Username `json:"username"`
}

type LoginResponse struct { // components/schemas/LoginResponse
	Identifier UserID `json:"identifier"`
}

type SetMyUsernameRequest struct { // components/schemas/SetMyUsernameRequest
	Username Username `json:"username"`
}

type SetPhotoRequest struct { // components/schemas/SetPhotoRequest
	PhotoURL PhotoURL `json:"photoUrl"`
}

type UserIdRequest struct { // components/schemas/UserIdRequest
	UserID UserID `json:"userId"`
}

type CreateGroupRequest struct { // components/schemas/CreateGroupRequest
	GroupName GroupName `json:"groupName"`
	MemberIds []UserID  `json:"memberIds"`
}

type SetGroupNameRequest struct { // components/schemas/SetGroupNameRequest
	Name GroupName `json:"name"`
}

type SendMessageRequest struct { // components/schemas/SendMessageRequest
	ReplyToMsgId InternalID      `json:"replyToMsgId,omitempty"`
	Text         *MessageContent `json:"text,omitempty"`
	PhotoURL     *PhotoURL       `json:"photoUrl,omitempty"`
}

type ForwardMessageRequest struct { // components/schemas/ForwardMessageRequest
	OriginalMessageId InternalID `json:"originalMessageId"`
}

type CommentMessageRequest struct { // components/schemas/CommentMessageRequest
	Emoji Emoji `json:"emoji"`
}
