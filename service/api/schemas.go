package api

import (
	"errors"
	"net/url"
	"regexp"
)

var (
	uuidRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
)

// -- Tipi Base e Validazioni --

type UserID string

func (id UserID) Validate() error {
	if !uuidRegex.MatchString(string(id)) {
		return errors.New("Formato UserID non valido. Deve essere un UUID.")
	}
	return nil
}

type InternalID string

func (id InternalID) Validate() error {
	if len(id) < 1 || len(id) > 50 {
		return errors.New("ID interno non valido")
	}
	return nil
}

type Username string

func (u Username) Validate() error {
	if len(u) < 3 || len(u) > 16 {
		return errors.New("Nome utente non valido (3-16 caratteri)")
	}
	return nil
}

type PhotoURL string

func (p PhotoURL) Validate() error {
	if len(p) < 1 || len(p) > 1000000 {
		return errors.New("URL foto troppo lungo o vuoto")
	}
	str := string(p)
	if len(str) > 5 && str[:5] == "data:" {
		return nil
	}
	if _, err := url.ParseRequestURI(str); err != nil {
		return errors.New("URL foto non valido")
	}
	return nil
}

type Emoji string

func (e Emoji) Validate() error {
	if len(e) < 1 || len(e) > 8 {
		return errors.New("Emoji non valido")
	}
	return nil
}

type Timestamp string

type GroupName string

func (g GroupName) Validate() error {
	if len(g) < 1 || len(g) > 50 {
		return errors.New("Nome gruppo non valido")
	}
	return nil
}

type MessageContent string

func (m MessageContent) Validate() error {
	if len(m) < 1 || len(m) > 4000 {
		return errors.New("Testo messaggio troppo lungo")
	}
	return nil
}

type SearchQuery string

func (s SearchQuery) Validate() error {
	if len(s) < 1 || len(s) > 16 {
		return errors.New("Query ricerca non valida")
	}
	return nil
}

// -- Schemi Request/Response --

type ErrorResponse struct {
	Message string `json:"message"`
}

type DoLoginRequest struct {
	Username Username `json:"username"`
}
type DoLoginResponse struct {
	Identifier UserID `json:"identifier"`
}

type SetMyUserNameRequest struct {
	Username Username `json:"username"`
}
type SetPhotoRequest struct {
	PhotoURL PhotoURL `json:"photoUrl"`
}
type UserIdRequest struct {
	UserID UserID `json:"userId"`
}

type CreateGroupRequest struct {
	GroupName GroupName `json:"groupName"`
	MemberIds []UserID  `json:"memberIds"`
}
type SetGroupNameRequest struct {
	Name GroupName `json:"name"`
}

// SendMessageRequest AGGIORNATA: Niente puntatori OneOf complessi.
type SendMessageRequest struct {
	ReplyToMsgId InternalID      `json:"replyToMsgId,omitempty"`
	Text         *MessageContent `json:"text,omitempty"`     // Puntatore per capire se è presente
	PhotoURL     *PhotoURL       `json:"photoUrl,omitempty"` // Puntatore per capire se è presente
}

type ForwardMessageRequest struct {
	OriginalMessageId InternalID `json:"originalMessageId"`
}
type CommentMessageRequest struct {
	Emoji Emoji `json:"emoji"`
}
