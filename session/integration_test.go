package session

import (
	"context"
	"testing"
	"time"

	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/kamilwrzyszcz/go_example/util"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

// Integration tests

type RedisIntegrationTestSuite struct {
	suite.Suite
	client *RedisClient
}

func TestRedisIntegrationTestSuite(t *testing.T) {
	suite.Run(t, &RedisIntegrationTestSuite{})
}

func (its *RedisIntegrationTestSuite) SetupSuite() {
	config, err := util.LoadConfig("./..")
	if err != nil {
		its.FailNowf("cannot load config: ", err.Error())
	}

	its.client, err = NewRedisClient(config.RedisAddress, config.RedisPassword)
	if err != nil {
		its.FailNowf("cannot create redis client: ", err.Error())
	}
}

func (its *RedisIntegrationTestSuite) AfterTest() {
	its.T().Log("i am ran after test!")
	cmd := its.client.rdb.FlushDB(context.Background())
	if err := cmd.Err(); err != nil {
		its.FailNowf("failed to flush redis: ", err.Error())
	}
}

func (its *RedisIntegrationTestSuite) TearDownSuite() {
	tearDownRedis(its)
}

// Tests

func (its *RedisIntegrationTestSuite) TestSetAndGet() {
	// Happy path
	session1 := createRandomSession()
	err := its.client.Set(context.Background(), session1.ID, session1)
	its.NoError(err)

	session2, err := its.client.Get(context.Background(), session1.ID)
	its.NoError(err)
	its.Equal(session1.ID, session2.ID)
	its.Equal(session1.Username, session2.Username)
	its.Equal(session1.RefreshToken, session2.RefreshToken)
	its.WithinDuration(session1.ExpiresAt, session2.ExpiresAt, time.Second)

	// Non-existing key
	session3, err := its.client.Get(context.Background(), "non-existing-key")
	its.Empty(session3)
	its.Error(err)

	// Fetching expired key
	session4 := createRandomSession()
	err = its.client.Set(context.Background(), session4.ID, session4)
	its.NoError(err)
	cmdBool := its.client.rdb.Expire(context.Background(), session4.ID, time.Second*0)
	its.True(cmdBool.Result())
	its.NoError(cmdBool.Err())

	session5, err := its.client.Get(context.Background(), session4.ID)
	its.Empty(session5)
	its.Error(err)
}

func (its *RedisIntegrationTestSuite) TestDelete() {
	session1 := createRandomSession()
	err := its.client.Set(context.Background(), session1.ID, session1)
	its.NoError(err)

	err = its.client.Del(context.Background(), session1.ID)
	its.NoError(err)

	session2, err := its.client.Get(context.Background(), session1.ID)
	its.Empty(session2)
	its.Error(err)

	err = its.client.Del(context.Background(), session1.ID)
	its.NoError(err)
}

// Setup helper functions

func createRandomSession() *Session {
	return &Session{
		ID:           util.RandomString(6),
		Username:     util.RandomAuthor(),
		RefreshToken: util.RandomString(25),
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Minute * 15),
		UserAgent:    "Mozilla",
		ClientIP:     "0.0.0.0",
		IsBlocked:    false,
	}
}

func tearDownRedis(its *RedisIntegrationTestSuite) {
	its.T().Log("tearing down redis")

	cmd := its.client.rdb.FlushDB(context.Background())
	if err := cmd.Err(); err != nil {
		its.FailNowf("failed to flush redis: ", err.Error())
	}

	err := its.client.rdb.Close()
	if err != nil {
		its.FailNowf("failed to close redis client: ", err.Error())
	}
}
