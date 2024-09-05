package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewConnection(databaseURL string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	return conn, nil
}
