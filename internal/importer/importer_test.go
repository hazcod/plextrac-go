package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cresco/plextrac-go/internal/testutil"
	"github.com/cresco/plextrac-go/pkg/plextrac"
)

func TestRun_DryRun(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "f.json")
	if err := os.WriteFile(path, []byte(`{"findings":[
		{"id":"A-1","title":"Finding A","severity":"Critical","description":"desc","recommendations":"fix"}
	]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	c, _ := plextrac.New("https://example.com", plextrac.WithToken("t"))
	var buf bytes.Buffer
	err := Run(context.Background(), Config{
		Client: c, Path: path, ClientID: "1", ReportID: "2",
		Mode: "upsert", DryRun: true, Out: &buf,
	})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Finding A") || !strings.Contains(out, "DRY RUN") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestRun_EndToEndCreate(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	s.On("POST", "/api/v1/client/1/report/2/flaw/create", func(w http.ResponseWriter, r *http.Request) {
		var f plextrac.Flaw
		_ = json.NewDecoder(r.Body).Decode(&f)
		f.ID = "new"
		testutil.JSON(w, 200, f)
	})
	c, _ := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("t"))

	tmp := t.TempDir()
	path := filepath.Join(tmp, "f.json")
	if err := os.WriteFile(path, []byte(`{"findings":[
		{"id":"A-1","title":"A","severity":"Alta","description":"d","recommendations":"r"}
	]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := Run(context.Background(), Config{
		Client: c, Path: path, ClientID: "1", ReportID: "2",
		Mode: "create", Out: &buf, Workers: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "CREATE A-1") {
		t.Fatalf("missing create log: %s", buf.String())
	}
	// severity Alta (Spanish) must be mapped to English "High"
	reqs := s.Requests()
	if len(reqs) != 1 {
		t.Fatalf("want 1 req got %d", len(reqs))
	}
	if !strings.Contains(string(reqs[0].Body), `"severity":"High"`) {
		t.Fatalf("severity not mapped: %s", reqs[0].Body)
	}
}
