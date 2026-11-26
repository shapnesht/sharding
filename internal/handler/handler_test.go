package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shapnesht/sharding/internal/store"
)

type fakeLocator struct {
	server string
	keys   *[]string
}

func (f fakeLocator) Locate(key string) string {
	*f.keys = append(*f.keys, key)
	return f.server
}

type insertCall struct{ server, url, urlID string }

type fakeStore struct {
	inserted  []insertCall
	insertErr error
	findCalls []insertCall
	ctxs      []context.Context
	findURL   string
	findID    string
	findErr   error
}

func (f *fakeStore) Insert(ctx context.Context, server, url, urlID string) error {
	f.ctxs = append(f.ctxs, ctx)
	f.inserted = append(f.inserted, insertCall{server, url, urlID})
	return f.insertErr
}

func (f *fakeStore) Find(ctx context.Context, server, urlID string) (string, string, error) {
	f.ctxs = append(f.ctxs, ctx)
	f.findCalls = append(f.findCalls, insertCall{server: server, urlID: urlID})
	return f.findURL, f.findID, f.findErr
}

func do(t *testing.T, s *fakeStore, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec, _ := doKeys(t, s, httptest.NewRequest(method, target, nil))
	return rec
}

// doKeys also returns the keys the handler asked the locator for.
func doKeys(t *testing.T, s *fakeStore, req *http.Request) (*httptest.ResponseRecorder, []string) {
	t.Helper()
	var keys []string
	rec := httptest.NewRecorder()
	New(fakeLocator{server: "5434", keys: &keys}, s).ServeHTTP(rec, req)
	return rec, keys
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) Response {
	t.Helper()
	var r Response
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("body is not JSON: %q", rec.Body.String())
	}
	return r
}

func TestPost_ReturnsShortUrlAndServer(t *testing.T) {
	s := &fakeStore{}
	rec := do(t, s, http.MethodPost, "/?url=https://example.com")

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content type %q", ct)
	}
	want := Response{Url: "https://example.com", ShortUrl: "10068", Server: "5434"}
	if got := decode(t, rec); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if len(s.inserted) != 1 || s.inserted[0] != (insertCall{"5434", "https://example.com", "10068"}) {
		t.Fatalf("insert calls: %+v", s.inserted)
	}
}

func TestPost_InsertError_Returns500(t *testing.T) {
	// TODO(bug): the handler keeps going after the error and also writes the JSON body.
	rec := do(t, &fakeStore{insertErr: errors.New("boom")}, http.MethodPost, "/?url=x")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "failed to save to DB") {
		t.Fatalf("body %q", rec.Body.String())
	}
}

func TestGet_ReturnsStoredUrl(t *testing.T) {
	s := &fakeStore{findURL: "https://example.com", findID: "10068"}
	rec := do(t, s, http.MethodGet, "/10068")

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	want := Response{Url: "https://example.com", ShortUrl: "10068", Server: "5434"}
	if got := decode(t, rec); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestGet_NotFound_Returns404(t *testing.T) {
	rec := do(t, &fakeStore{findErr: store.ErrNotFound}, http.MethodGet, "/zzzzz")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Short url doesn't exists") {
		t.Fatalf("body %q", rec.Body.String())
	}
}

func TestGet_StoreError_Returns500(t *testing.T) {
	rec := do(t, &fakeStore{findErr: errors.New("boom")}, http.MethodGet, "/10068")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "error finding url from DB") {
		t.Fatalf("body %q", rec.Body.String())
	}
}

func TestOtherMethods_Return405(t *testing.T) {
	rec := do(t, &fakeStore{}, http.MethodPut, "/")

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Method not allowed") {
		t.Fatalf("body %q", rec.Body.String())
	}
}

func TestGet_LocatesAndFindsByPathWithoutSlash(t *testing.T) {
	s := &fakeStore{findURL: "u", findID: "10068"}
	_, keys := doKeys(t, s, httptest.NewRequest(http.MethodGet, "/10068", nil))

	if len(keys) != 1 || keys[0] != "10068" {
		t.Fatalf("locator keys: %v", keys)
	}
	if len(s.findCalls) != 1 || s.findCalls[0] != (insertCall{server: "5434", urlID: "10068"}) {
		t.Fatalf("find calls: %+v", s.findCalls)
	}
}

func TestPost_LocatesByShortID(t *testing.T) {
	_, keys := doKeys(t, &fakeStore{}, httptest.NewRequest(http.MethodPost, "/?url=https://example.com", nil))

	if len(keys) != 1 || keys[0] != "10068" {
		t.Fatalf("locator keys: %v", keys)
	}
}

// A client disconnect must not cancel a DB call: pgx would close the shard's only connection.
func TestDBCalls_IgnoreRequestCancellation(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s := &fakeStore{}
		doKeys(t, s, httptest.NewRequest(method, "/10068?url=x", nil).WithContext(ctx))

		if len(s.ctxs) != 1 || s.ctxs[0].Err() != nil {
			t.Fatalf("%s: store got a cancelled context", method)
		}
	}
}
