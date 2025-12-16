package database

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhotoURL string `json:"photoUrl,omitempty"`
}

type ConversationSummary struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	PhotoURL               string `json:"photoUrl,omitempty"`
	LatestMessageSnippet   string `json:"latestMessageSnippet,omitempty"`
	LatestMessageTimestamp string `json:"latestMessageTimestamp,omitempty"`
}

// components/schemas/Conversation
type Conversation struct {
	ID                     string    `json:"id"`
	Name                   string    `json:"name"`
	PhotoURL               string    `json:"photoUrl,omitempty"`
	IsGroup                bool      `json:"isGroup"`
	LatestMessageSnippet   string    `json:"latestMessageSnippet,omitempty"`   // <--- AGGIUNTO
	LatestMessageTimestamp string    `json:"latestMessageTimestamp,omitempty"` // <--- AGGIUNTO
	Members                []User    `json:"members"`
	Messages               []Message `json:"messages"`
}

// Struttura Messaggio AGGIORNATA
type Message struct {
	ID           string     `json:"id"`
	Sender       User       `json:"sender"`
	Text         string     `json:"text,omitempty"`     // Testo opzionale
	PhotoURL     string     `json:"photoUrl,omitempty"` // Foto opzionale
	Timestamp    string     `json:"timestamp"`
	Status       string     `json:"status,omitempty"`
	ReplyToMsgId *string    `json:"replyToMsgId,omitempty"`
	Reactions    []Reaction `json:"reactions"`
}

type Reaction struct {
	ID    string `json:"id"`
	Emoji string `json:"emoji"`
	User  User   `json:"user"`
}

type UserList struct {
	Users []User `json:"users"`
}

type ConversationList struct {
	Conversations []ConversationSummary `json:"conversations"`
}
