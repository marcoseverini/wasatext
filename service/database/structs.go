package database

// components/schemas/User
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhotoURL string `json:"photoUrl,omitempty"`
}

// components/schemas/ConversationSummary
type ConversationSummary struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	PhotoURL               string `json:"photoUrl,omitempty"`
	LatestMessageSnippet   string `json:"latestMessageSnippet,omitempty"`
	LatestMessageTimestamp string `json:"latestMessageTimestamp,omitempty"`
}

// components/schemas/Conversation
type Conversation struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	PhotoURL string    `json:"photoUrl,omitempty"`
	IsGroup  bool      `json:"isGroup"`
	Members  []User    `json:"members"`
	Messages []Message `json:"messages"`
}

// components/schemas/Message
type Message struct {
	ID           string     `json:"id"`
	Sender       User       `json:"sender"`
	Content      string     `json:"content"`
	ContentType  string     `json:"contentType"`
	Timestamp    string     `json:"timestamp"`
	Status       string     `json:"status,omitempty"`
	ReplyToMsgId *string    `json:"replyToMsgId,omitempty"`
	Reactions    []Reaction `json:"reactions"`
}

// components/schemas/Reaction
type Reaction struct {
	ID    string `json:"id"`
	Emoji string `json:"emoji"`
	User  User   `json:"user"`
}

// components/schemas/UserList
type UserList struct {
	Users []User `json:"users"`
}

// components/schemas/ConversationList
type ConversationList struct {
	Conversations []ConversationSummary `json:"conversations"`
}
