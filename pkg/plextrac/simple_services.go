package plextrac

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// TemplateService covers report/findings templates.
type TemplateService struct{ c *Client }

func (s *TemplateService) List(ctx context.Context, kind string, opts ListOpts) *Iter[Template] {
	return newIter[Template](ctx, func(ctx context.Context, page int) ([]Template, bool, error) {
		o := opts
		if o.Page == 0 {
			o.Page = page
		}
		if o.Extra == nil {
			o.Extra = map[string][]string{}
		}
		if kind != "" {
			o.Extra.Set("kind", kind)
		}
		path := "/api/v1/templates"
		if q := o.query().Encode(); q != "" {
			path += "?" + q
		}
		var resp struct {
			Templates []Template `json:"templates"`
			Data      []Template `json:"data"`
			HasMore   bool       `json:"has_more"`
		}
		if err := s.c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return nil, false, err
		}
		items := resp.Templates
		if items == nil {
			items = resp.Data
		}
		return items, resp.HasMore && opts.Page == 0, nil
	})
}

func (s *TemplateService) Get(ctx context.Context, templateID string) (*Template, error) {
	if err := validateID(templateID); err != nil {
		return nil, err
	}
	var out Template
	if err := s.c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/template/%s", templateID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// TagService covers /tags.
type TagService struct{ c *Client }

func (s *TagService) List(ctx context.Context) ([]Tag, error) {
	var resp struct {
		Tags []Tag `json:"tags"`
		Data []Tag `json:"data"`
	}
	if err := s.c.do(ctx, http.MethodGet, "/api/v1/tags", nil, &resp); err != nil {
		return nil, err
	}
	if resp.Tags != nil {
		return resp.Tags, nil
	}
	return resp.Data, nil
}

func (s *TagService) Create(ctx context.Context, t Tag) (*Tag, error) {
	var out Tag
	if err := s.c.do(ctx, http.MethodPost, "/api/v1/tag/create", t, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UserService covers tenant user management.
type UserService struct{ c *Client }

func (s *UserService) List(ctx context.Context, opts ListOpts) *Iter[User] {
	return newIter[User](ctx, func(ctx context.Context, page int) ([]User, bool, error) {
		o := opts
		if o.Page == 0 {
			o.Page = page
		}
		path := "/api/v1/users"
		if q := o.query().Encode(); q != "" {
			path += "?" + q
		}
		var resp struct {
			Users   []User `json:"users"`
			Data    []User `json:"data"`
			HasMore bool   `json:"has_more"`
		}
		if err := s.c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return nil, false, err
		}
		items := resp.Users
		if items == nil {
			items = resp.Data
		}
		return items, resp.HasMore && opts.Page == 0, nil
	})
}

func (s *UserService) Get(ctx context.Context, userID string) (*User, error) {
	if err := validateID(userID); err != nil {
		return nil, err
	}
	var out User
	if err := s.c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/user/%s", userID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ExportService starts and polls report exports (PDF, DOCX, etc).
type ExportService struct{ c *Client }

func (s *ExportService) Start(ctx context.Context, clientID, reportID, format string) (*Export, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	body := map[string]string{"format": format}
	var out Export
	path := fmt.Sprintf("/api/v1/client/%s/report/%s/export", clientID, reportID)
	if err := s.c.do(ctx, http.MethodPost, path, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ExportService) Get(ctx context.Context, clientID, reportID, exportID string) (*Export, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	if err := validateID(exportID); err != nil {
		return nil, err
	}
	var out Export
	path := fmt.Sprintf("/api/v1/client/%s/report/%s/export/%s", clientID, reportID, exportID)
	if err := s.c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// jsonUnmarshal is a small helper so other files can avoid re-importing
// encoding/json when they only need this one call.
func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }
