package plextrac

import (
	"context"
	"fmt"
	"net/http"
)

// WriteupService covers content-library writeup endpoints.
type WriteupService struct{ c *Client }

func (s *WriteupService) List(ctx context.Context, opts ListOpts) *Iter[Writeup] {
	return newIter[Writeup](ctx, func(ctx context.Context, page int) ([]Writeup, bool, error) {
		o := opts
		if o.Page == 0 {
			o.Page = page
		}
		path := "/api/v1/writeups"
		if q := o.query().Encode(); q != "" {
			path += "?" + q
		}
		var resp struct {
			Writeups []Writeup `json:"writeups"`
			Data     []Writeup `json:"data"`
			HasMore  bool      `json:"has_more"`
		}
		if err := s.c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return nil, false, err
		}
		items := resp.Writeups
		if items == nil {
			items = resp.Data
		}
		return items, resp.HasMore && opts.Page == 0, nil
	})
}

func (s *WriteupService) Get(ctx context.Context, writeupID string) (*Writeup, error) {
	if err := validateID(writeupID); err != nil {
		return nil, err
	}
	var out Writeup
	if err := s.c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/writeup/%s", writeupID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Apply materialises a writeup into a report as a new flaw.
func (s *WriteupService) Apply(ctx context.Context, clientID, reportID, writeupID string) (*Flaw, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	if err := validateID(writeupID); err != nil {
		return nil, err
	}
	var out Flaw
	path := fmt.Sprintf("/api/v1/client/%s/report/%s/writeup/%s/apply", clientID, reportID, writeupID)
	if err := s.c.do(ctx, http.MethodPost, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
