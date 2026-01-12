package api

import (
	"errors"   // Libreria per gestire gli errori
	"net/http" // Libreria per gestire HTTP

	"github.com/julienschmidt/httprouter"                // Router HTTP di terze parti
	"github.com/marcoseverini/wasatext/service/database" // Database
	"github.com/sirupsen/logrus"                         // Libreria di logging strutturato
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

	// Crea il router HTTP di terze parti
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	// Crea l'istanza di _router (rt) salvando le dipendenze
	rt := &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}

	// ROTTE API (Backend)
	router.POST("/session", rt.doLogin)
	router.PUT("/settings/username", rt.authMiddleware(rt.setMyUserName))
	router.PUT("/settings/photo", rt.authMiddleware(rt.setMyPhoto))
	router.GET("/users", rt.authMiddleware(rt.searchUsers))
	router.POST("/conversations", rt.authMiddleware(rt.startConversation))
	router.GET("/conversations", rt.authMiddleware(rt.getMyConversations))
	router.POST("/conversations/:convId/messages", rt.authMiddleware(rt.sendMessage))
	router.GET("/conversations/:convId", rt.authMiddleware(rt.getConversation))
	router.DELETE("/messages/:msgId", rt.authMiddleware(rt.deleteMessage))
	router.POST("/conversations/:convId/forwarded_messages", rt.authMiddleware(rt.forwardMessage))
	router.POST("/messages/:msgId/reactions", rt.authMiddleware(rt.commentMessage))
	router.DELETE("/messages/:msgId/reactions/:reactionId", rt.authMiddleware(rt.uncommentMessage))
	router.POST("/groups", rt.authMiddleware(rt.createGroup))
	router.PUT("/conversations/:convId/name", rt.authMiddleware(rt.setGroupName))
	router.PUT("/conversations/:convId/photo", rt.authMiddleware(rt.setGroupPhoto))
	router.POST("/conversations/:convId/members", rt.authMiddleware(rt.addToGroup))
	router.DELETE("/conversations/:convId/members/me", rt.authMiddleware(rt.leaveGroup))

	// ROTTE STATICHE (Frontend)
	// Queste servono i file generati da 'yarn build' nella cartella webui/dist

	// Serve la Index (Home Page)
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.ServeFile(w, r, "webui/dist/index.html")
	})

	// Serve i file Javascript e CSS (cartella assets)
	router.ServeFiles("/assets/*filepath", http.Dir("webui/dist/assets"))

	// Serve Bootstrap (che hai in public/bootstrap e finisce in dist/bootstrap)
	router.ServeFiles("/bootstrap/*filepath", http.Dir("webui/dist/bootstrap"))

	// Serve l'icona SVG (feather sprite)
	router.GET("/feather-sprite-v4.29.0.svg", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.ServeFile(w, r, "webui/dist/feather-sprite-v4.29.0.svg")
	})

	// Serve la Favicon (opzionale)
	router.GET("/favicon.ico", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.ServeFile(w, r, "webui/dist/favicon.ico")
	})

	return rt, nil
}

// Ritorna il gestore HTTP pronto per il server
func (rt *_router) Handler() http.Handler {
	return rt.router
}

func (rt *_router) Close() error {
	return nil
}
