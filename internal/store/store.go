// Package store holds the SQL for the Urls table.
package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned by Find when no row matches.
var ErrNotFound = errors.New("short url not found")

// Store runs queries against one connection per shard, keyed by shard name.
type Store struct {
	conns map[string]*pgx.Conn
}

func New(conns map[string]*pgx.Conn) *Store {
	return &Store{conns: conns}
}

// Insert saves a url under its short id on the given shard.
func (s *Store) Insert(ctx context.Context, server, url, urlID string) error {
	conn := s.conns[server]
	conn.Prepare(ctx, "insertNewUrl", "INSERT INTO Urls(url, url_id) VALUES($1, $2)")
	// TODO(bug): Query is used for an INSERT and the rows are never closed; Exec is the right call.
	_, err := conn.Query(ctx, "insertNewUrl", url, urlID)
	return err
}

// Find returns the url and short id stored for urlID on the given shard.
func (s *Store) Find(ctx context.Context, server, urlID string) (url, shortID string, err error) {
	conn := s.conns[server]
	conn.Prepare(ctx, "queryShortUrl", "SELECT url, url_id FROM Urls WHERE url_id = $1")
	err = conn.QueryRow(ctx, "queryShortUrl", urlID).Scan(&url, &shortID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return url, shortID, err
}
