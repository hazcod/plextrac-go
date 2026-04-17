package plextrac

import (
	"context"
	"fmt"
	"net/http"
)

// ReportService covers /client/{clientID}/report* endpoints.
type ReportService struct{ c *Client }

func (s *ReportService) List(ctx context.Context, clientID string, opts ListOpts) *Iter[Report] {
	return newIter[Report](ctx, func(ctx context.Context, page int) ([]Report, bool, error) {
		if err := validateID(clientID); err != nil {
			return nil, false, err
		}
		o := opts
		if o.Page == 0 {
			o.Page = page
		}
		path := fmt.Sprintf("/api/v1/client/%s/reports", clientID)
		if q := o.query().Encode(); q != "" {
			path += "?" + q
		}
		var resp struct {
			Reports []Report `json:"reports"`
			Data    []Report `json:"data"`
			HasMore bool     `json:"has_more"`
		}
		if err := s.c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return nil, false, err
		}
		items := resp.Reports
		if items == nil {
			items = resp.Data
		}
		return items, resp.HasMore && opts.Page == 0, nil
	})
}

func (s *ReportService) Get(ctx context.Context, clientID, reportID string) (*Report, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	var out Report
	if err := s.c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/client/%s/report/%s", clientID, reportID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ReportService) Create(ctx context.Context, clientID string, r *Report) (*Report, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	var out Report
	if err := s.c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/client/%s/report/create", clientID), r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ReportService) Update(ctx context.Context, clientID, reportID string, r *Report) (*Report, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	var out Report
	if err := s.c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/client/%s/report/%s", clientID, reportID), r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ReportService) Delete(ctx context.Context, clientID, reportID string) error {
	if err := validateID(clientID); err != nil {
		return err
	}
	if err := validateID(reportID); err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/client/%s/report/%s", clientID, reportID), nil, nil)
}
