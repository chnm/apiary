package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/orlangure/gnomock"
	"github.com/orlangure/gnomock/preset/postgres"
	"github.com/stretchr/testify/require"
)

func TestDBConnection(t *testing.T) {
	t.Parallel()

	user := "gnomock"
	pass := "strong-passwords-are-the-best"
	dbname := "gnomock_test"

	p := postgres.Preset(
		postgres.WithUser(user, pass),
		postgres.WithDatabase(dbname),
	)

	container, err := gnomock.Start(p)
	require.NoError(t, err)
	defer func() { require.NoError(t, gnomock.Stop(container)) }()

	connstr := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?sslmode=disable",
		user, pass, container.Host, container.DefaultPort(), dbname)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := Connect(ctx, connstr)
	require.NoError(t, err)

	err = pool.Ping(ctx)
	require.NoError(t, err)

	require.IsType(t, &pgxpool.Pool{}, pool)
}
