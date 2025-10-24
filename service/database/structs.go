package database

// Struttura che rappresenta un utente nel database
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
