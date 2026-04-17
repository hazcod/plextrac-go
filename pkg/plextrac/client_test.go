package plextrac_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/cresco/plextrac-go/internal/testutil"
	"github.com/cresco/plextrac-go/pkg/plextrac"
)

func TestNew_RejectsNonHTTPS(t *testing.T) {
	t.Parallel()
	_, err := plextrac.New("http://insecure")
	if !errors.Is(err, plextrac.ErrInvalidBaseURL) {
		t.Fatalf("want ErrInvalidBaseURL, got %v", err)
	}
}

func TestNew_AllowHTTP(t *testing.T) {
	t.Parallel()
	_, err := plextrac.New("http://localhost:9999", plextrac.WithAllowHTTP())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_InvalidURL(t *testing.T) {
	t.Parallel()
	_, err := plextrac.New(":::not a url")
	if !errors.Is(err, plextrac.ErrInvalidBaseURL) {
		t.Fatalf("want ErrInvalidBaseURL, got %v", err)
	}
}

func TestDo_RetriesOn5xx(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	attempts := 0
	s.On("GET", "/api/v1/clients", func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			http.Error(w, "boom", http.StatusServiceUnavailable)
			return
		}
		testutil.JSON(w, 200, map[string]any{"clients": []map[string]string{{"name": "ok"}}})
	})
	c, err := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("x"),
		plextrac.WithRetry(5, 0))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	out, err := c.Clients.List(ctx, plextrac.ListOpts{}).All(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(out) != 1 || out[0].Name != "ok" {
		t.Fatalf("unexpected: %+v", out)
	}
	if attempts != 3 {
		t.Fatalf("want 3 attempts, got %d", attempts)
	}
}

func TestDo_SurfacesAPIError(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	s.On("GET", "/api/v1/clients", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, http.StatusNotFound, map[string]string{"message": "nope"})
	})
	c, _ := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("x"),
		plextrac.WithRetry(0, 0))
	_, err := c.Clients.List(context.Background(), plextrac.ListOpts{}).All(context.Background())
	var apiErr *plextrac.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("want APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Fatalf("want 404, got %d", apiErr.StatusCode)
	}
	if !errors.Is(err, plextrac.ErrNotFound) {
		t.Fatalf("errors.Is ErrNotFound failed")
	}
}

func TestDo_SendsBearerToken(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	s.On("GET", "/api/v1/clients", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, map[string]any{"clients": []any{}})
	})
	c, _ := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("secret-jwt"))
	_, _ = c.Clients.List(context.Background(), plextrac.ListOpts{}).All(context.Background())
	reqs := s.Requests()
	if len(reqs) == 0 || reqs[0].Auth != "Bearer secret-jwt" {
		t.Fatalf("want Bearer header, got %+v", reqs)
	}
}
