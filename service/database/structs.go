package database

// Struttura che rappresenta un utente nel database
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	PhotoURL string `json:"photoUrl,omitempty"`
}

// ConversationSummary rappresenta l'anteprima di una conversazione.
type ConversationSummary struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	PhotoURL               string `json:"photoUrl,omitempty"`
	LatestMessageSnippet   string `json:"latestMessageSnippet,omitempty"`
	LatestMessageTimestamp string `json:"latestMessageTimestamp,omitempty"`
}

// Conversation rappresenta una conversazione completa.
type Conversation struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	PhotoURL string    `json:"photoUrl,omitempty"`
	IsGroup  bool      `json:"isGroup"`
	Members  []User    `json:"members"`
	Messages []Message `json:"messages"`
}

// Message rappresenta un singolo messaggio.
type Message struct {
	ID          string     `json:"id"`
	Sender      User       `json:"sender"` 
	Content     string     `json:"content"`
	ContentType string     `json:"contentType"`
	Timestamp   string     `json:"timestamp"`
	Status      string     `json:"status,omitempty"`
	Reactions   []Reaction `json:"reactions"`
}

// Reaction rappresenta una reazione.
type Reaction struct {
	ID    string `json:"id"`
	Emoji string `json:"emoji"`
	User  User   `json:"user"`
}