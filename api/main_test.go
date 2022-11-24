package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/kamilwrzyszcz/go_example/db/sqlc"
	"github.com/kamilwrzyszcz/go_example/session"
	"github.com/kamilwrzyszcz/go_example/util"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, store db.Store, sessionClient session.SessionClient) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store, sessionClient)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
