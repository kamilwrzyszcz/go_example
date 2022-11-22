package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/kamilwrzyszcz/go_example/db/sqlc"
	"github.com/kamilwrzyszcz/go_example/session"
	"github.com/kamilwrzyszcz/go_example/token"
	"github.com/kamilwrzyszcz/go_example/util"
)

// Server serves HTTP requests for our banking service
type Server struct {
	config        util.Config
	store         db.Store
	sessionClient session.SessionClient
	tokenMaker    token.Maker
	router        *gin.Engine
}

// NewServer creates a new HTTP server and setup routing
func NewServer(config util.Config, store db.Store, sessionClient session.SessionClient) (*Server, error) {
	tokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:        config,
		store:         store,
		sessionClient: sessionClient,
		tokenMaker:    tokenMaker,
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	// TODO: progressively add routes

	server.router = router
}

// Start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// Reusable error response func
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
