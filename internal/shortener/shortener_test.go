package shortener

import "testing"

func TestShortID_KnownVector(t *testing.T) {
	if got := ShortID("https://example.com"); got != "10068" {
		t.Fatalf("got %q, want 10068", got)
	}
}

func TestShortID_EmptyURL(t *testing.T) {
	if got := ShortID(""); got != "e3b0c" {
		t.Fatalf("got %q, want e3b0c", got)
	}
}

func TestShortID_Deterministic(t *testing.T) {
	if ShortID("a") != ShortID("a") {
		t.Fatal("same input produced different ids")
	}
}
