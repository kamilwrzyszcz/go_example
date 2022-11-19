package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/kamilwrzyszcz/go_example/util"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
)

// Integration tests

type IntegrationTestSuite struct {
	suite.Suite
	db          *sql.DB
	testQueries *Queries
	m           *migrate.Migrate
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, &IntegrationTestSuite{})
}

func (its *IntegrationTestSuite) SetupSuite() {
	config, err := util.LoadConfig("../..")
	if err != nil {
		its.FailNowf("cannot load config: ", err.Error())
	}

	testDB, err := sql.Open(config.DBDriver, config.DBTestSource)
	if err != nil {
		its.FailNowf("cannot connect to the db: ", err.Error())
	}
	testQueries := New(testDB)

	its.db = testDB
	its.testQueries = testQueries

	setupDatabase(its)
}

func (its *IntegrationTestSuite) TearDownSuite() {
	tearDownDatabase(its)
}

// Tests

func createRandomArticle(its *IntegrationTestSuite) Article {
	user := createRandomUser(its)

	arg := CreateArticleParams{
		Author:   user.Username,
		Headline: util.RandomString(15),
		Content:  util.RandomString(25),
	}

	article, err := its.testQueries.CreateArticle(context.Background(), arg)
	its.NoError(err)
	its.NotEmpty(article)

	its.Equal(arg.Author, article.Author)
	its.Equal(arg.Headline, article.Headline)
	its.Equal(arg.Content, article.Content)

	its.NotZero(article.ID)
	its.NotZero(article.CreatedAt)

	return article
}

func createRandomUser(its *IntegrationTestSuite) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	its.NoError(err)

	arg := CreateUserParams{
		Username:       util.RandomAuthor(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomAuthor(),
		Email:          util.RandomEmail(),
	}

	user, err := its.testQueries.CreateUser(context.Background(), arg)
	its.NoError(err)
	its.NotEmpty(user)

	its.Equal(arg.Username, user.Username)
	its.Equal(arg.HashedPassword, user.HashedPassword)
	its.Equal(arg.FullName, user.FullName)
	its.Equal(arg.Email, user.Email)

	its.True(user.PasswordChangedAt.IsZero())
	its.NotZero(user.CreatedAt)

	return user
}

func (its *IntegrationTestSuite) TestCreateUser() {
	createRandomUser(its)
}

func (its *IntegrationTestSuite) TestGetUser() {
	user1 := createRandomUser(its)
	user2, err := its.testQueries.GetUser(context.Background(), user1.Username)
	its.NoError(err)
	its.NotEmpty(user2)

	its.Equal(user1.Username, user2.Username)
	its.Equal(user1.HashedPassword, user2.HashedPassword)
	its.Equal(user1.FullName, user2.FullName)
	its.Equal(user1.Email, user2.Email)
	its.WithinDuration(user1.CreatedAt, user2.CreatedAt, time.Second)
	its.WithinDuration(user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second)

	user3, err := its.testQueries.GetUser(context.Background(), "non-existing-user")
	its.Error(err)
	its.ErrorIs(err, sql.ErrNoRows)
	its.Empty(user3)
}

func (its *IntegrationTestSuite) TestCreateArticle() {
	createRandomArticle(its)
}

func (its *IntegrationTestSuite) TestGetArticle() {
	article1 := createRandomArticle(its)
	article2, err := its.testQueries.GetArticle(context.Background(), article1.ID)
	its.NoError(err)
	its.NotEmpty(article2)

	its.Equal(article1.ID, article2.ID)
	its.Equal(article1.Author, article2.Author)
	its.Equal(article1.Headline, article2.Headline)
	its.Equal(article1.Content, article2.Content)
	its.WithinDuration(article1.CreatedAt, article2.CreatedAt, time.Second)

	article3, err := its.testQueries.GetArticle(context.Background(), 2137)
	its.Error(err)
	its.EqualError(err, sql.ErrNoRows.Error())
	its.Empty(article3)
}

func (its *IntegrationTestSuite) TestUpdateArticle() {
	article1 := createRandomArticle(its)

	arg1 := UpdateArticleParams{
		Headline: sql.NullString{
			String: "updated headline",
			Valid:  true,
		},
		Content: sql.NullString{
			String: "updated content",
			Valid:  true,
		},
		ID: article1.ID,
	}

	article2, err := its.testQueries.UpdateArticle(context.Background(), arg1)
	its.NoError(err)
	its.NotEmpty(article2)

	its.Equal(article1.ID, article2.ID)
	its.Equal(article1.Author, article2.Author)
	its.Equal(arg1.Headline.String, article2.Headline)
	its.Equal(arg1.Content.String, article2.Content)
	its.WithinDuration(article1.CreatedAt, article2.CreatedAt, time.Second)
	its.WithinDuration(article2.EditedAt.Time, time.Now(), time.Second)

	arg2 := UpdateArticleParams{
		Headline: sql.NullString{
			String: "updated headline 2",
			Valid:  true,
		},
		ID: article1.ID,
	}

	article3, err := its.testQueries.UpdateArticle(context.Background(), arg2)
	its.NoError(err)
	its.NotEmpty(article3)

	its.Equal(article1.ID, article3.ID)
	its.Equal(article1.Author, article3.Author)
	its.Equal(arg2.Headline.String, article3.Headline)
	its.Equal(article2.Content, article3.Content)
	its.WithinDuration(article1.CreatedAt, article3.CreatedAt, time.Second)
	its.WithinDuration(article3.EditedAt.Time, time.Now(), time.Second)
}

func (its *IntegrationTestSuite) TestDeleteArticle() {
	article1 := createRandomArticle(its)
	err := its.testQueries.DeleteArticle(context.Background(), article1.ID)
	its.NoError(err)

	article2, err := its.testQueries.GetArticle(context.Background(), article1.ID)
	its.Error(err)
	its.EqualError(err, sql.ErrNoRows.Error())
	its.Empty(article2)
}

func (its *IntegrationTestSuite) TestListArticles() {
	var lastArticle Article
	for i := 0; i < 10; i++ {
		lastArticle = createRandomArticle(its)
	}

	arg := ListArticlesParams{
		Author: lastArticle.Author,
		Limit:  5,
		Offset: 0,
	}

	articles, err := its.testQueries.ListArticles(context.Background(), arg)
	its.NoError(err)
	its.NotEmpty(articles)

	for _, article := range articles {
		its.NotEmpty(article)
		its.Equal(lastArticle.Author, article.Author)
	}
}

// Setup helper functions

func setupDatabase(its *IntegrationTestSuite) {
	its.T().Log("setting up database")

	driver, err := postgres.WithInstance(its.db, &postgres.Config{})
	if err != nil {
		its.FailNowf("cannot get driver: ", err.Error())
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://../migration",
		"test_db", driver)
	if err != nil {
		its.FailNowf("cannot get migration files: ", err.Error())
	}

	err = m.Up()
	if err != nil {
		its.FailNowf("failed to migrate up: ", err.Error())
	}

	its.m = m
}

func tearDownDatabase(its *IntegrationTestSuite) {
	its.T().Log("tearing down database")

	err := its.m.Down()
	if err != nil {
		its.FailNowf("failed to migrate down: ", err.Error())
	}
}
