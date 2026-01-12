package api

import (
	"errors"
	"net/url"
	"regexp"
)

var (
	uuidRegex = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)
)

type UserID string

func (id UserID) Validate() error {
	if !uuidRegex.MatchString(string(id)) {
		return errors.New("Formato UserID non valido.")
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
		return errors.New("Username non valido")
	}
	return nil
}

type PhotoURL string

func (p PhotoURL) Validate() error {
	if len(p) < 1 || len(p) > 1000000 {
		return errors.New("URL foto non valido")
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
		return errors.New("Emoji non valida")
	}
	return nil
}

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
		return errors.New("Testo troppo lungo")
	}
	return nil
}

type SearchQuery string

func (s SearchQuery) Validate() error {
	if len(s) < 1 || len(s) > 16 {
		return errors.New("Query non valida")
	}
	return nil
}

// Structs
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

type SendMessageRequest struct {
	ReplyToMsgId InternalID      `json:"replyToMsgId,omitempty"`
	Text         *MessageContent `json:"text,omitempty"`
	PhotoURL     *PhotoURL       `json:"photoUrl,omitempty"`
}

type ForwardMessageRequest struct {
	OriginalMessageId InternalID `json:"originalMessageId"`
}

type CommentMessageRequest struct {
	Emoji Emoji `json:"emoji"`
}
