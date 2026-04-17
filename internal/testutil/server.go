// Package testutil provides a fake Plextrac HTTP server for SDK tests.
package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// Handler is a test-supplied request handler keyed by METHOD + path.
type Handler func(w http.ResponseWriter, r *http.Request)

// Server is a scripted fake Plextrac. Register per-route handlers via On().
type Server struct {
	t        *testing.T
	mu       sync.Mutex
	ts       *httptest.Server
	routes   map[string]Handler
	requests []RecordedRequest
}

// RecordedRequest captures one request for test assertions.
type RecordedRequest struct {
	Method string
	Path   string
	Body   []byte
	Auth   string
}

// New starts a fake server. Call Close on cleanup.
func New(t *testing.T) *Server {
	t.Helper()
	s := &Server{t: t, routes: map[string]Handler{}}
	s.ts = httptest.NewServer(http.HandlerFunc(s.dispatch))
	t.Cleanup(func() { s.ts.Close() })
	return s
}

// URL returns the base URL.
func (s *Server) URL() string { return s.ts.URL }

// On registers a handler for the given method + path.
func (s *Server) On(method, path string, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.routes[method+" "+path] = h
}

// Requests returns a copy of the recorded requests.
func (s *Server) Requests() []RecordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]RecordedRequest, len(s.requests))
	copy(out, s.requests)
	return out
}

func (s *Server) dispatch(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	s.mu.Lock()
	s.requests = append(s.requests, RecordedRequest{
		Method: r.Method, Path: r.URL.Path, Body: body, Auth: r.Header.Get("Authorization"),
	})
	h, ok := s.routes[r.Method+" "+r.URL.Path]
	s.mu.Unlock()
	if !ok {
		http.Error(w, "not found: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		return
	}
	// Replay body so handlers can read it
	r.Body = io.NopCloser(newBuffer(body))
	h(w, r)
}

// JSON is a helper that writes a JSON response.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// newBuffer avoids importing bytes in tests.
func newBuffer(b []byte) *buf { return &buf{b: b} }

type buf struct {
	b []byte
	i int
}

func (b *buf) Read(p []byte) (int, error) {
	if b.i >= len(b.b) {
		return 0, io.EOF
	}
	n := copy(p, b.b[b.i:])
	b.i += n
	return n, nil
}

func (b *buf) Close() error { return nil }
