package shard

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Connect opens a connection to one shard.
func Connect(ctx context.Context, host, dbName, username, password string, port int16) (*pgx.Conn, error) {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, host, port, dbName)
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
