package api

import (
	"errors" // Libreria per gestire gli errori
	"net/http" // Strumenti per gestire l'HTTP
	"github.com/marcoseverini/wasatext/service/database" // Il nostro database
	"github.com/julienschmidt/httprouter" // router HTTP di terze parti
	"github.com/sirupsen/logrus" // Libreria di logging strutturato
)

// Dipendenze per il router API
type Config struct {
	Logger   logrus.FieldLogger
	Database database.AppDatabase
}

// Interfaccia del router API
type Router interface {
	Handler() http.Handler
	Close() error
}

// Implementazione concreta del router API
type _router struct {
	router *httprouter.Router
	
	baseLogger logrus.FieldLogger
	db         database.AppDatabase 
}

// Costruttore del router API
func New(cfg Config) (Router, error) {

	// Validazione delle dipendenze
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database is required")
	}

	// Creiamo il router HTTP di terze parti
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	// Creiamo l'istanza di _router (rt) salvando le dipendenze (db e logger) al suo interno.
	rt := &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}

	// Rotte HTTP
	router.POST("/session", rt.doLogin)
	router.PUT("/settings/username", rt.authMiddleware(rt.setMyUserName))
	router.PUT("/settings/photo", rt.authMiddleware(rt.setMyPhoto))
	
	
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