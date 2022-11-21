package session

import (
	"context"
	"time"
)

// Session struct is used to handle sessions
type Session struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserAgent    string    `json:"user_agent"`
	ClientIP     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
}

type SessionClient interface {
	Set(ctx context.Context, key string, session *Session) error
	Get(ctx context.Context, key string) (*Session, error)
	Del(ctx context.Context, key string) error
}
