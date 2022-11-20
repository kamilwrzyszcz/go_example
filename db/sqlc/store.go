package db

import (
	"database/sql"
)

// Store provides all functions to execute db queries
type Store interface {
	Querier
}

// SQLStore provides all functions to execute SQL queries
type SQLStore struct {
	*Queries
	db *sql.DB
}

// NewStore creates and returns a new Store
func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}
