package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jackc/pgx/v4/stdlib"
)

// Connect returns a pool of connection to the database using the pgx native interface.
// It also opens a separate connection using the standard SQL interface for compatibility with
// previous handlers.
func Connect(ctx context.Context, connstr string) (*pgxpool.Pool, *sql.DB, error) {

	var pool *pgxpool.Pool

	cfg, err := pgxpool.ParseConfig(connstr)
	if err != nil {
		return nil, nil, err
	}

	connectWithRetry := func() error {
		select {
		case <-ctx.Done():
			return backoff.Permanent(errors.New("Cancelled attempt to connect to database"))
		default:
			conn, err := pgxpool.ConnectConfig(ctx, cfg)
			if err != nil {
				return fmt.Errorf("Error creating connection pool: %w", err)
			}

			err = conn.Ping(ctx)
			if err != nil {
				return fmt.Errorf("Error pinging database: %w", err)
			}
			pool = conn
		}
		return nil
	}

	policy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)
	err = backoff.Retry(connectWithRetry, policy)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to connect to database: %w", err)
	}

	connStr := stdlib.RegisterConnConfig(cfg.ConnConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to connect to database: %w", err)
	}

	return pool, db, nil
}
