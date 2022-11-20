package redis

import "time"

// Session struct is used to handle sessions
type Session struct {
	ID           string
	Username     string
	RefreshToken string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	UserAgent    string
	ClientIP     string
	IsBlocked    bool
}
