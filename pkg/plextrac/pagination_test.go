package plextrac_test

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/cresco/plextrac-go/internal/testutil"
	"github.com/cresco/plextrac-go/pkg/plextrac"
)

func TestIter_FetchesPagesLazily(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	pagesSeen := 0
	s.On("GET", "/api/v1/clients", func(w http.ResponseWriter, r *http.Request) {
		pagesSeen++
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		switch page {
		case 1:
			testutil.JSON(w, 200, map[string]any{
				"data":     []plextrac.ClientOrg{{Name: "a"}, {Name: "b"}},
				"has_more": true,
			})
		case 2:
			testutil.JSON(w, 200, map[string]any{
				"data":     []plextrac.ClientOrg{{Name: "c"}},
				"has_more": false,
			})
		default:
			http.Error(w, "bad page", 400)
		}
	})
	c, _ := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("t"))
	iter := c.Clients.List(context.Background(), plextrac.ListOpts{})
	all, err := iter.All(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("want 3 items, got %d: %+v", len(all), all)
	}
	if pagesSeen != 2 {
		t.Fatalf("want 2 page fetches, got %d", pagesSeen)
	}
}
