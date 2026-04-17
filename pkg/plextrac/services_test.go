package plextrac_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/hazcod/plextrac-go/internal/testutil"
	"github.com/hazcod/plextrac-go/pkg/plextrac"
)

func newTestClient(t *testing.T) (*plextrac.Client, *testutil.Server) {
	t.Helper()
	s := testutil.New(t)
	c, err := plextrac.New(s.URL(), plextrac.WithAllowHTTP(), plextrac.WithToken("t"),
		plextrac.WithRetry(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	return c, s
}

func TestClients_Full(t *testing.T) {
	t.Parallel()
	c, s := newTestClient(t)
	s.On("GET", "/api/v1/client/42", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.ClientOrg{ID: "42", Name: "Acme"})
	})
	s.On("POST", "/api/v1/client/create", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.ClientOrg{ID: "new", Name: "Acme"})
	})
	s.On("PUT", "/api/v1/client/42", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.ClientOrg{ID: "42", Name: "Acme2"})
	})
	s.On("DELETE", "/api/v1/client/42", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(204)
	})
	ctx := context.Background()

	got, err := c.Clients.Get(ctx, "42")
	if err != nil || got.Name != "Acme" {
		t.Fatalf("get: %+v %v", got, err)
	}
	created, err := c.Clients.Create(ctx, &plextrac.ClientOrg{Name: "Acme"})
	if err != nil || created.ID != "new" {
		t.Fatalf("create: %+v %v", created, err)
	}
	updated, err := c.Clients.Update(ctx, "42", &plextrac.ClientOrg{Name: "Acme2"})
	if err != nil || updated.Name != "Acme2" {
		t.Fatalf("update: %+v %v", updated, err)
	}
	if err := c.Clients.Delete(ctx, "42"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestReports_Full(t *testing.T) {
	t.Parallel()
	c, s := newTestClient(t)
	s.On("GET", "/api/v1/client/1/reports", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, map[string]any{"reports": []plextrac.Report{{ID: "r1", Name: "R"}}})
	})
	s.On("GET", "/api/v1/client/1/report/r1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Report{ID: "r1", Name: "R"})
	})
	s.On("POST", "/api/v1/client/1/report/create", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Report{ID: "new", Name: "R"})
	})
	s.On("PUT", "/api/v1/client/1/report/r1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Report{ID: "r1", Name: "R2"})
	})
	s.On("DELETE", "/api/v1/client/1/report/r1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(204)
	})
	ctx := context.Background()
	all, err := c.Reports.List(ctx, "1", plextrac.ListOpts{}).All(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("list: %+v %v", all, err)
	}
	if _, err := c.Reports.Get(ctx, "1", "r1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Reports.Create(ctx, "1", &plextrac.Report{Name: "R"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Reports.Update(ctx, "1", "r1", &plextrac.Report{Name: "R2"}); err != nil {
		t.Fatal(err)
	}
	if err := c.Reports.Delete(ctx, "1", "r1"); err != nil {
		t.Fatal(err)
	}
}

func TestAssets_Full(t *testing.T) {
	t.Parallel()
	c, s := newTestClient(t)
	s.On("GET", "/api/v1/client/1/assets", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, map[string]any{"assets": []plextrac.Asset{{ID: "a1", Name: "host"}}})
	})
	s.On("GET", "/api/v1/client/1/asset/a1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Asset{ID: "a1", Name: "host"})
	})
	s.On("POST", "/api/v1/client/1/asset/create", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Asset{ID: "new", Name: "host"})
	})
	s.On("PUT", "/api/v1/client/1/asset/a1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Asset{ID: "a1", Name: "host2"})
	})
	s.On("DELETE", "/api/v1/client/1/asset/a1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(204)
	})
	s.On("PUT", "/api/v1/client/1/report/r1/asset/a1", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(204)
	})
	ctx := context.Background()
	if _, err := c.Assets.List(ctx, "1", plextrac.ListOpts{}).All(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Assets.Get(ctx, "1", "a1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Assets.Create(ctx, "1", &plextrac.Asset{Name: "host"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Assets.Update(ctx, "1", "a1", &plextrac.Asset{Name: "host2"}); err != nil {
		t.Fatal(err)
	}
	if err := c.Assets.Delete(ctx, "1", "a1"); err != nil {
		t.Fatal(err)
	}
	if err := c.Assets.AttachToReport(ctx, "1", "r1", "a1"); err != nil {
		t.Fatal(err)
	}
}

func TestWriteups_Full(t *testing.T) {
	t.Parallel()
	c, s := newTestClient(t)
	s.On("GET", "/api/v1/writeups", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, map[string]any{"writeups": []plextrac.Writeup{{ID: "w1", Title: "WU"}}})
	})
	s.On("GET", "/api/v1/writeup/w1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Writeup{ID: "w1", Title: "WU"})
	})
	s.On("POST", "/api/v1/client/1/report/r1/writeup/w1/apply", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Flaw{ID: "applied", Title: "WU"})
	})
	ctx := context.Background()
	if _, err := c.Writeups.List(ctx, plextrac.ListOpts{}).All(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Writeups.Get(ctx, "w1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Writeups.Apply(ctx, "1", "r1", "w1"); err != nil {
		t.Fatal(err)
	}
}

func TestTemplates_And_Tags_Users(t *testing.T) {
	t.Parallel()
	c, s := newTestClient(t)
	s.On("GET", "/api/v1/templates", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, map[string]any{"templates": []plextrac.Template{{ID: "t1", Name: "T"}}})
	})
	s.On("GET", "/api/v1/template/t1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Template{ID: "t1", Name: "T"})
	})
	s.On("GET", "/api/v1/tags", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, map[string]any{"tags": []plextrac.Tag{{Name: "web"}}})
	})
	s.On("POST", "/api/v1/tag/create", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Tag{Name: "web"})
	})
	s.On("GET", "/api/v1/users", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, map[string]any{"users": []plextrac.User{{ID: "u1", Email: "a@b"}}})
	})
	s.On("GET", "/api/v1/user/u1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.User{ID: "u1", Email: "a@b"})
	})
	ctx := context.Background()
	if _, err := c.Templates.List(ctx, "report", plextrac.ListOpts{}).All(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Templates.Get(ctx, "t1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Tags.List(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Tags.Create(ctx, plextrac.Tag{Name: "web"}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Users.List(ctx, plextrac.ListOpts{}).All(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Users.Get(ctx, "u1"); err != nil {
		t.Fatal(err)
	}
}

func TestExports_StartGet(t *testing.T) {
	t.Parallel()
	c, s := newTestClient(t)
	s.On("POST", "/api/v1/client/1/report/r1/export", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Export{ID: "e1", Format: "pdf", Status: "queued"})
	})
	s.On("GET", "/api/v1/client/1/report/r1/export/e1", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, plextrac.Export{ID: "e1", Status: "done", URL: "https://x"})
	})
	ctx := context.Background()
	e, err := c.Exports.Start(ctx, "1", "r1", "pdf")
	if err != nil || e.Status != "queued" {
		t.Fatalf("start: %+v %v", e, err)
	}
	got, err := c.Exports.Get(ctx, "1", "r1", "e1")
	if err != nil || got.URL != "https://x" {
		t.Fatalf("get: %+v %v", got, err)
	}
}

func TestAttachments_Upload(t *testing.T) {
	t.Parallel()
	c, s := newTestClient(t)
	s.On("POST", "/api/v1/client/1/report/r1/flaw/f1/upload", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("wrong content-type: %s", r.Header.Get("Content-Type"))
		}
		testutil.JSON(w, 200, plextrac.Attachment{ID: "att1", Filename: "x.png", Size: 3})
	})
	att, err := c.Attachments.Upload(context.Background(), "1", "r1", "f1", "x.png", bytes.NewReader([]byte("png")))
	if err != nil {
		t.Fatal(err)
	}
	if att.ID != "att1" {
		t.Fatalf("got %+v", att)
	}
}

func TestAuth_PasswordMFAFlow(t *testing.T) {
	t.Parallel()
	s := testutil.New(t)
	// first call: password login returns MFA challenge
	s.On("POST", "/api/v1/authenticate", func(w http.ResponseWriter, _ *http.Request) {
		testutil.JSON(w, 200, map[string]any{
			"status": "success", "mfa_enabled": true, "code": "challenge-abc",
		})
	})
	s.On("POST", "/api/v1/authenticate/mfa", func(w http.ResponseWriter, r *http.Request) {
		var got map[string]string
		_ = json.NewDecoder(r.Body).Decode(&got)
		if got["code"] != "challenge-abc" || got["token"] != "123456" {
			t.Errorf("wrong MFA body: %+v", got)
		}
		testutil.JSON(w, 200, map[string]string{"token": "jwt-xyz"})
	})
	s.On("GET", "/api/v1/clients", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer jwt-xyz" {
			t.Errorf("wrong auth: %s", r.Header.Get("Authorization"))
		}
		testutil.JSON(w, 200, map[string]any{"clients": []any{}})
	})
	c, err := plextrac.New(s.URL(),
		plextrac.WithAllowHTTP(),
		plextrac.WithPasswordAuth("alice", "pass", plextrac.StaticMFA("123456")),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Clients.List(context.Background(), plextrac.ListOpts{}).All(context.Background()); err != nil {
		t.Fatal(err)
	}
}
