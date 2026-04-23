package postgres

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Credentials contains postgres connection credentials.
type Credentials struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBName     string
	DBPort     string
}

// New creates a postgres connection pool.
func New(ctx context.Context, creds Credentials) (*pgxpool.Pool, error) {
	connectionURI := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		creds.DBUser,
		creds.DBPassword,
		net.JoinHostPort(creds.DBHost, creds.DBPort),
		creds.DBName,
	)

	log.Printf(
		"Connecting to postgres database with name '%s' at %s:%s",
		creds.DBName,
		creds.DBHost,
		creds.DBPort,
	)

	pool, err := pgxpool.New(ctx, connectionURI)
	if err != nil {
		return pool, fmt.Errorf("%w: failed to create pool", err)
	}

	return pool, nil
}
