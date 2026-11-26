package shard

import "testing"

var shardNames = []string{"5433", "5434", "5435", "5436"}

func TestRing_LocatesKnownKeys(t *testing.T) {
	r := NewRing(shardNames)
	want := map[string]string{
		"10068": "5435",
		"e3b0c": "5435",
		"a1b2c": "5434",
		"00000": "5436",
	}
	for key, shard := range want {
		if got := r.Locate(key); got != shard {
			t.Errorf("Locate(%q) = %q, want %q", key, got, shard)
		}
	}
}

func TestRing_Deterministic(t *testing.T) {
	r := NewRing(shardNames)
	first := r.Locate("10068")
	for i := 0; i < 100; i++ {
		if got := r.Locate("10068"); got != first {
			t.Fatalf("call %d: got %q, want %q", i, got, first)
		}
	}
}
