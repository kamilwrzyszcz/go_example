package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/kamilwrzyszcz/go_example/session"
	mockSession "github.com/kamilwrzyszcz/go_example/session/mock"
	"github.com/kamilwrzyszcz/go_example/token"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker token.Maker,
	sessionID uuid.UUID,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, payload, err := tokenMaker.CreateToken(sessionID, username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
	sessionID, err := uuid.NewRandom()
	require.NoError(t, err)
	username := "user"

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID)
		buildStubs    func(sessionClient *mockSession.MockSessionClient)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs: func(sessionClient *mockSession.MockSessionClient) {
				arg := &session.Session{
					ID:           sessionID.String(),
					Username:     username,
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
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {

			},
			buildStubs: func(sessionClient *mockSession.MockSessionClient) {
				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, "unsupported", "user", time.Minute)
			},
			buildStubs: func(sessionClient *mockSession.MockSessionClient) {
				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, "", "user", time.Minute)
			},
			buildStubs: func(sessionClient *mockSession.MockSessionClient) {
				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, "user", -time.Minute)
			},
			buildStubs: func(sessionClient *mockSession.MockSessionClient) {
				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "NoSession",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs: func(sessionClient *mockSession.MockSessionClient) {
				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&session.Session{}, redis.Nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SessionBlocked",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs: func(sessionClient *mockSession.MockSessionClient) {
				arg := &session.Session{
					ID:           sessionID.String(),
					Username:     username,
					RefreshToken: "refresh_token",
					CreatedAt:    time.Now(),
					ExpiresAt:    time.Now().Add(time.Hour * 24),
					UserAgent:    "Mozilla",
					ClientIP:     "0.0.0.0",
					IsBlocked:    true,
				}

				sessionClient.EXPECT().
					Get(gomock.Any(), gomock.Eq(sessionID.String())).
					Times(1).
					Return(arg, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "SessionMismatch",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker, sessionID uuid.UUID) {
				addAuthorization(t, request, tokenMaker, sessionID, authorizationTypeBearer, "user", time.Minute)
			},
			buildStubs: func(sessionClient *mockSession.MockSessionClient) {
				arg := &session.Session{
					ID:           sessionID.String(),
					Username:     "user2",
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
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sessionClient := mockSession.NewMockSessionClient(ctrl)
			tc.buildStubs(sessionClient)

			server := newTestServer(t, nil, sessionClient)

			authPath := "/fake"
			server.router.GET(
				authPath,
				authMiddleware(server.tokenMaker, sessionClient),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker, sessionID)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
