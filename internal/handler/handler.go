// Package handler serves the URL shortener HTTP API.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shapnesht/sharding/internal/shortener"
	"github.com/shapnesht/sharding/internal/store"
)

type Response struct {
	Url      string `json:"url"`
	ShortUrl string `json:"shortUrl"`
	Server   string `json:"server"`
}

// Locator picks the shard that owns a key.
type Locator interface {
	Locate(key string) string
}

// URLStore saves and looks up urls on a given shard.
type URLStore interface {
	Insert(ctx context.Context, server, url, urlID string) error
	Find(ctx context.Context, server, urlID string) (url, shortID string, err error)
}

type handler struct {
	locator Locator
	store   URLStore
}

// New returns the HTTP handler: GET looks a short url up, POST creates one.
func New(l Locator, s URLStore) http.Handler {
	h := &handler{locator: l, store: s}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.getURL(w, r)
		case http.MethodPost:
			h.postURL(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}

func (h *handler) getURL(w http.ResponseWriter, r *http.Request) {
	urlID := r.URL.Path[1:]
	server := h.locator.Locate(urlID)
	url, shortID, err := h.store.Find(context.Background(), server, urlID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.Error(w, `{"error":"Short url doesn't exists"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"error finding url from DB"}`, http.StatusInternalServerError)
		}
		return
	}
	resp := Response{Url: url, ShortUrl: shortID, Server: server}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"error":"failed to encode JSON"}`, http.StatusInternalServerError)
	}
}

func (h *handler) postURL(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	shortURL := shortener.ShortID(url)
	w.Header().Set("Content-Type", "application/json")
	server := h.locator.Locate(shortURL)
	if err := h.store.Insert(context.Background(), server, url, shortURL); err != nil {
		// TODO(bug): execution continues and the JSON body is written after the error.
		http.Error(w, `{"error":"failed to save to DB"}`, http.StatusInternalServerError)
	}
	resp := Response{Url: url, ShortUrl: shortURL, Server: server}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"error":"failed to encode JSON"}`, http.StatusInternalServerError)
	}
}
