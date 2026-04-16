package utils

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	secrets "github.com/DomNidy/saint_sim/internal/secrets"
)

// UUIDString converts a Postgres UUID to its string representation.
func UUIDString(id pgtype.UUID) (string, error) {
	value, err := id.Value()
	if err != nil {
		return "", fmt.Errorf("read uuid value: %w", err)
	}

	uuidString, ok := value.(string)
	if !ok {
		return "", errUnexpectedUUIDValueType
	}

	return uuidString, nil
}

// InitPostgresConnectionPool creates a postgres connection pool.
func InitPostgresConnectionPool(ctx context.Context) (*pgxpool.Pool, error) {
	dbUser := secrets.LoadSecret("DB_USER").Value()
	dbPassword := secrets.LoadSecret("DB_PASSWORD").Value()
	dbHost := secrets.LoadSecret("DB_HOST").Value()
	dbName := secrets.LoadSecret("DB_NAME").Value()
	dbPort := "5432"

	connectionURI := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		dbUser,
		dbPassword,
		net.JoinHostPort(dbHost, dbPort),
		dbName,
	)

	log.Printf("Connecting to postgres database with name '%s' at %s:%s", dbName, dbHost, dbPort)

	pool, err := pgxpool.New(ctx, connectionURI)
	if err != nil {
		return pool, fmt.Errorf("%w: failed to create pool", err)
	}

	return pool, nil
}
