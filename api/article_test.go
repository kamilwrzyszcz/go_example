package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	mockdb "github.com/kamilwrzyszcz/go_example/db/mock"
	db "github.com/kamilwrzyszcz/go_example/db/sqlc"
	"github.com/kamilwrzyszcz/go_example/session"
	mockSession "github.com/kamilwrzyszcz/go_example/session/mock"
	"github.com/kamilwrzyszcz/go_example/token"
	"github.com/kamilwrzyszcz/go_example/util"
	"github.com/stretchr/testify/require"
)

// Could write much more but it shows general idea

func TestArticleAPI(t *testing.T) {
	user, _ := randomUser(t)
	article := randomArticle(user.Username)
	sessionID, err := uuid.NewRandom()
	require.NoError(t, err)

	testCases := []struct {
		name          string
		url           string
		httpMethod    string
		body          gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID)
		buildStubs    func(store *mockdb.MockStore, sessionClient *mockSession.MockSessionClient)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:       "OK",
			url:        fmt.Sprintf("/articles/%d", article.ID),
			httpMethod: http.MethodGet,
			body:       nil,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, sessionClient *mockSession.MockSessionClient) {
				arg := &session.Session{
					ID:           sessionID.String(),
					Username:     user.Username,
					RefreshToken: "refresh_token",
					CreatedAt:    time.Now(),
					ExpiresAt:    time.Now().Add(time.Hour * 24),
					UserAgent:    "Mozilla",
					ClientIP:     "0.0.0.0",
					IsBlocked:    false,
				}

				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Eq(sessionID.String())).
					Times(1).
					Return(arg, nil)

				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Eq(article.ID)).
					Times(1).
					Return(article, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchArticle(t, recorder.Body, article)
			},
		},
		{
			name:       "NotFound",
			url:        fmt.Sprintf("/articles/%d", 2137),
			httpMethod: http.MethodGet,
			body:       nil,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, sessionClient *mockSession.MockSessionClient) {
				arg := &session.Session{
					ID:           sessionID.String(),
					Username:     user.Username,
					RefreshToken: "refresh_token",
					CreatedAt:    time.Now(),
					ExpiresAt:    time.Now().Add(time.Hour * 24),
					UserAgent:    "Mozilla",
					ClientIP:     "0.0.0.0",
					IsBlocked:    false,
				}

				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Eq(sessionID.String())).
					Times(1).
					Return(arg, nil)

				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Article{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:       "OKList",
			url:        "/articles?page_id=2&page_size=5",
			httpMethod: http.MethodGet,
			body:       nil,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, sessionClient *mockSession.MockSessionClient) {
				arg_session := &session.Session{
					ID:           sessionID.String(),
					Username:     user.Username,
					RefreshToken: "refresh_token",
					CreatedAt:    time.Now(),
					ExpiresAt:    time.Now().Add(time.Hour * 24),
					UserAgent:    "Mozilla",
					ClientIP:     "0.0.0.0",
					IsBlocked:    false,
				}

				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Eq(sessionID.String())).
					Times(1).
					Return(arg_session, nil)

				arg_store := db.ListArticlesParams{
					Author: article.Author,
					Limit:  5,
					Offset: (2 - 1) * 5,
				}
				articles := []db.Article{article, article}
				store.EXPECT().
					ListArticles(gomock.Any(), gomock.Eq(arg_store)).
					Times(1).
					Return(articles, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:       "Unauthorized",
			url:        fmt.Sprintf("/articles/%d", article.ID),
			httpMethod: http.MethodGet,
			body:       nil,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, sessionClient *mockSession.MockSessionClient) {
				arg := &session.Session{
					ID:           sessionID.String(),
					Username:     user.Username,
					RefreshToken: "refresh_token",
					CreatedAt:    time.Now(),
					ExpiresAt:    time.Now().Add(time.Hour * 24),
					UserAgent:    "Mozilla",
					ClientIP:     "0.0.0.0",
					IsBlocked:    false,
				}

				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Eq(sessionID.String())).
					Times(1).
					Return(arg, nil)

				store.EXPECT().
					GetArticle(gomock.Any(), gomock.Eq(article.ID)).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name:       "Created",
			url:        "/articles",
			httpMethod: http.MethodPost,
			body: gin.H{
				"headline": article.Headline,
				"content":  article.Content,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore, sessionClient *mockSession.MockSessionClient) {
				arg_session := &session.Session{
					ID:           sessionID.String(),
					Username:     user.Username,
					RefreshToken: "refresh_token",
					CreatedAt:    time.Now(),
					ExpiresAt:    time.Now().Add(time.Hour * 24),
					UserAgent:    "Mozilla",
					ClientIP:     "0.0.0.0",
					IsBlocked:    false,
				}

				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Eq(sessionID.String())).
					Times(1).
					Return(arg_session, nil)

				arg_store := db.CreateArticleParams{
					Author:   user.Username,
					Headline: article.Headline,
					Content:  article.Content,
				}
				store.EXPECT().
					CreateArticle(gomock.Any(), gomock.Eq(arg_store)).
					Times(1).
					Return(article, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
				requireBodyMatchArticle(t, recorder.Body, article)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			sessionClient := mockSession.NewMockSessionClient(ctrl)
			tc.buildStubs(store, sessionClient)

			// start test server and send request
			server := newTestServer(t, store, sessionClient)
			recorder := httptest.NewRecorder()

			// Marshal body data to JSON
			var request *http.Request
			if tc.body != nil {
				body, err := json.Marshal(tc.body)
				require.NoError(t, err)

				request, err = http.NewRequest(tc.httpMethod, tc.url, bytes.NewReader(body))
				require.NoError(t, err)
			} else {
				request, err = http.NewRequest(tc.httpMethod, tc.url, nil)
				require.NoError(t, err)
			}

			tc.setupAuth(t, request, server.tokenMaker, sessionID)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

// Helper functions

func randomArticle(author string) db.Article {
	return db.Article{
		ID:       util.RandomInt(1, 1000),
		Author:   author,
		Headline: util.RandomString(10),
		Content:  util.RandomString(25),
	}
}

func requireBodyMatchArticle(t *testing.T, body *bytes.Buffer, article db.Article) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotArticle db.Article
	err = json.Unmarshal(data, &gotArticle)
	require.NoError(t, err)
	require.Equal(t, article, gotArticle)
}
