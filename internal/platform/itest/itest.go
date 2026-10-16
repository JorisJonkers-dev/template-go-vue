//go:build integration

// Package itest runs the whole service the way main assembles it, on a database of the test's own,
// and answers over real HTTP, so an integration test checks a behaviour the way a browser reaches it.
package itest

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JorisJonkers-dev/template-go-vue/internal/assembly"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/httpx"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg/pgtest"
	"github.com/JorisJonkers-dev/template-go-vue/internal/server"
)

// Service is the running service.
type Service struct {
	url string
}

// Start runs the service for the rest of the test.
func Start(t *testing.T) *Service {
	t.Helper()
	store, err := pg.Open(t.Context(), pgtest.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(store.Close)
	logger := slog.New(slog.DiscardHandler)
	routes, err := assembly.Routes(assembly.Config{Logger: logger, Store: store})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(server.New(logger, "test", time.Second, routes).Handler())
	t.Cleanup(srv.Close)
	return &Service{url: srv.URL}
}

// Call sends method and path, with body as JSON unless it is nil, as subject, or with no identity
// when subject is empty. It fails the test unless the answer has status want, and decodes the
// answer into out unless out is nil.
func (s *Service) Call(t *testing.T, subject, method, path string, body any, want int, out any) {
	t.Helper()
	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		payload = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(t.Context(), method, s.url+path, payload)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if subject != "" {
		req.Header.Set(httpx.IdentityHeader, subject)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != want {
		t.Fatalf("%s %s = %d %s, want %d", method, path, resp.StatusCode, raw, want)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("%s %s: %v in %s", method, path, err, raw)
		}
	}
}
