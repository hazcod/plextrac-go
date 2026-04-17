package plextrac_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/cresco/plextrac-go/internal/testutil"
	"github.com/cresco/plextrac-go/pkg/plextrac"
)

func TestFlaws_Get(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	s.On("GET", "/api/v1/client/1/report/2/flaw/abc", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Flaw{ID: "abc", Title: "T", Severity: plextrac.SeverityHigh})
	})
	c, _ := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("t"))
	f, err := c.Flaws.Get(context.Background(), "1", "2", "abc")
	if err != nil {
		t.Fatal(err)
	}
	if f.Title != "T" || f.Severity != plextrac.SeverityHigh {
		t.Fatalf("unexpected: %+v", f)
	}
}

func TestFlaws_InvalidIDRejected(t *testing.T) {
	t.Parallel()
	c, _ := plextrac.New("https://example.com", plextrac.WithToken("t"))
	_, err := c.Flaws.Get(context.Background(), "../evil", "2", "3")
	if !errors.Is(err, plextrac.ErrInvalidID) {
		t.Fatalf("want ErrInvalidID, got %v", err)
	}
}

func TestFlaws_CreateSendsExpectedBody(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	s.On("POST", "/api/v1/client/1/report/2/flaw/create", func(w http.ResponseWriter, r *http.Request) {
		var got plextrac.Flaw
		_ = json.NewDecoder(r.Body).Decode(&got)
		got.ID = "new-id"
		testutil.JSON(w, 200, got)
	})
	c, _ := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("t"))
	in := &plextrac.Flaw{
		Title:       "Finding",
		Severity:    plextrac.SeverityCritical,
		Status:      plextrac.StatusOpen,
		Description: "<b>desc</b>",
	}
	out, err := c.Flaws.Create(context.Background(), "1", "2", in)
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "new-id" {
		t.Fatalf("want new-id, got %q", out.ID)
	}
	reqs := s.Requests()
	if len(reqs) != 1 {
		t.Fatalf("want 1 req, got %d", len(reqs))
	}
	if !strings.Contains(string(reqs[0].Body), "<b>desc</b>") {
		t.Fatalf("body missing raw HTML: %s", reqs[0].Body)
	}
}

func TestFlaws_BatchUpsertParallel(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	s.On("POST", "/api/v1/client/1/report/2/flaw/create", func(w http.ResponseWriter, r *http.Request) {
		var got plextrac.Flaw
		_ = json.NewDecoder(r.Body).Decode(&got)
		got.ID = "created-" + got.Title
		testutil.JSON(w, 200, got)
	})
	s.On("PUT", "/api/v1/client/1/report/2/flaw/exists-id", func(w http.ResponseWriter, r *http.Request) {
		var got plextrac.Flaw
		_ = json.NewDecoder(r.Body).Decode(&got)
		got.ID = "exists-id"
		testutil.JSON(w, 200, got)
	})
	c, _ := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("t"),
		plextrac.WithConcurrency(4))
	flaws := []*plextrac.Flaw{
		{Title: "A"},
		{Title: "existing"},
		{Title: "B"},
	}
	idx := map[string]string{"existing": "exists-id"}
	results, err := c.Flaws.BatchUpsert(context.Background(), "1", "2", flaws, idx, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("want 3 results, got %d", len(results))
	}
	var create, update int
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("err at %d: %v", r.Index, r.Err)
		}
		if r.Action == "create" {
			create++
		} else {
			update++
		}
	}
	if create != 2 || update != 1 {
		t.Fatalf("want 2 creates + 1 update, got %d/%d", create, update)
	}
}
