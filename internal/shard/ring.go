// Package shard owns the consistent-hash ring and shard connections.
package shard

import (
	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash/v2"
)

type hasher struct{}

func (hasher) Sum64(data []byte) uint64 { return xxhash.Sum64(data) }

type member string

func (m member) String() string { return string(m) }

// Ring maps keys to shard names.
type Ring struct {
	c *consistent.Consistent
}

// NewRing builds a ring over the given shard names.
func NewRing(names []string) *Ring {
	c := consistent.New(nil, consistent.Config{
		PartitionCount:    7,
		ReplicationFactor: 1,
		Load:              1.25,
		Hasher:            hasher{},
	})
	for _, n := range names {
		c.Add(member(n))
	}
	return &Ring{c: c}
}

// Locate returns the name of the shard that owns key.
func (r *Ring) Locate(key string) string {
	return r.c.LocateKey([]byte(key)).String()
}
