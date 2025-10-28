package api

// il JSON in entrata dalla richiesta POST /session
// Corrisponde a #/components/schemas/LoginRequest
type LoginRequest struct {
	Name string `json:"name"`
}

// il JSON in uscita dalla risposta POST /session
// Corrisponde a #/components/schemas/LoginResponse
type LoginResponse struct {
	Identifier string `json:"identifier"`
}

// ErrorResponse è una struttura generica per gli errori JSON
// Corrisponde a #/components/schemas/Error
type ErrorResponse struct {
	Message string `json:"message"`
}

// SetUsernameRequest è la struttura per la richiesta di aggiornamento del nome utente
type SetUsernameRequest struct {
	Username string `json:"username"`
}

// SetPhotoRequest è la struttura per leggere il JSON da PUT /settings/photo
type SetPhotoRequest struct {
	PhotoURL string `json:"photoUrl"` // Corretto: Tag 'json', nome campo JSON 'photoUrl'
}
