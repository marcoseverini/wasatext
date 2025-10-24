package api

import (
	"errors"
	"net/http" // Strumenti per gestire l'HTTP

	"wasatext/service/database" // Il nostro database

	"github.com/julienschmidt/httprouter" // router HTTP di terze parti
	"github.com/sirupsen/logrus" // Libreria di logging strutturato
)

// Dipendenze (come il DB)
type Config struct {
	Logger   logrus.FieldLogger
	Database database.AppDatabase
}

// Interfaccia del pacchetto
type Router interface {
	Handler() http.Handler
	Close() error
}

// Implementazione concreta
type _router struct {
	router *httprouter.Router
	
	baseLogger logrus.FieldLogger
	db         database.AppDatabase 
}

// Costruttore dell'API (chiamato da main.go)
func New(cfg Config) (Router, error) {

	// Validazione delle dipendenze
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database is required")
	}

	// Creiamo il router HTTP
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	// Creiamo l'istanza di _router (rt)
	// salvando le dipendenze (db e logger) al suo interno.
	rt := &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}

	// Registriamo le rotte HTTP

	router.POST("/session", rt.doLogin)
	
	// Qui, in futuro, aggiungeremo tutte le altre rotte:
	// router.PUT("/settings/username", rt.setMyUserName)
	// router.GET("/conversations", rt.getMyConversations)
	// ...
	
	// Restituiamo il router configurato
	return rt, nil
}

// Handler ritorna il gestore HTTP pronto per il server
func (rt *_router) Handler() http.Handler {
	return rt.router
}

// Per ora non fa nulla, ma potrebbe chiudere connessioni in futuro
func (rt *_router) Close() error {
	return nil
}