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

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker, server.sessionClient))

	authRoutes.POST("/users/logout", server.logoutUser)

	authRoutes.POST("/articles", server.createArticle)
	authRoutes.GET("/articles/:id", server.getArticle)
	authRoutes.GET("/articles", server.listArticles)
	authRoutes.DELETE("/articles/:id", server.deleteArticle)
	authRoutes.PATCH("/articles/:id", server.updateArticle)

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
