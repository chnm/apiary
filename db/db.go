package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Connect returns a pool of connections to the database using the pgx native interface.
func Connect(ctx context.Context, connstr string) (*pgxpool.Pool, error) {

	var pool *pgxpool.Pool

	cfg, err := pgxpool.ParseConfig(connstr)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("Failed to connect to database: %w", err)
	}

	return pool, nil
}
