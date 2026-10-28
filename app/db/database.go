package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect create a connection with database
func Connect(database_url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), database_url)
	if err != nil {
		return nil, errors.New("could not connect to database")
	}
	return pool, nil
}
