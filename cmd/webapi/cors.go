package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{
			"content-type", "authorization", "x-example-header",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// ATTENZIONE: Questo '*' è fondamentale per i Docker separati
		handlers.AllowedOrigins([]string{"*"}),
	)(h)
}
