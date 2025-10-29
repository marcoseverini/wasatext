package api

// Struttura generica per gli errori JSON
// Corrisponde a #/components/schemas/Error
type ErrorResponse struct {
	Message string `json:"message"`
}

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

// Il JSON in entrata dalla richiesta PUT /settings/username
// Corrisponde a #/components/schemas/SetUsernameRequest
type SetUsernameRequest struct {
	Username string `json:"username"`
}

// Il JSON in entrata dalla richiesta PUT /settings/photo
// Corrisponde a #/components/schemas/SetPhotoRequest
type SetPhotoRequest struct {
	PhotoURL string `json:"photoUrl"` // Corretto: Tag 'json', nome campo JSON 'photoUrl'
}
