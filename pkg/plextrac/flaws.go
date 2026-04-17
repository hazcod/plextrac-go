package plextrac

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

// FlawService covers /client/{clientID}/report/{reportID}/flaw* endpoints.
type FlawService struct{ c *Client }

// List returns an iterator over the report's flaws. Each page is requested
// on demand; pass an Opts with Page=0 to start from the first page.
func (s *FlawService) List(ctx context.Context, clientID, reportID string, opts ListOpts) *Iter[Flaw] {
	fetch := func(ctx context.Context, page int) ([]Flaw, bool, error) {
		if err := validateID(clientID); err != nil {
			return nil, false, err
		}
		if err := validateID(reportID); err != nil {
			return nil, false, err
		}
		o := opts
		if o.Page == 0 {
			o.Page = page
		}
		path := fmt.Sprintf("/api/v1/client/%s/report/%s/flaws", clientID, reportID)
		q := o.query().Encode()
		if q != "" {
			path += "?" + q
		}
		var resp struct {
			Flaws   []Flaw `json:"flaws"`
			Data    []Flaw `json:"data"`
			Results []Flaw `json:"results"`
			HasMore bool   `json:"has_more"`
		}
		if err := s.c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return nil, false, err
		}
		items := resp.Flaws
		if items == nil {
			items = resp.Data
		}
		if items == nil {
			items = resp.Results
		}
		hasMore := resp.HasMore && len(items) > 0 && opts.Page == 0
		return items, hasMore, nil
	}
	return newIter[Flaw](ctx, fetch)
}

// Get fetches a single flaw.
func (s *FlawService) Get(ctx context.Context, clientID, reportID, flawID string) (*Flaw, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	if err := validateID(flawID); err != nil {
		return nil, err
	}
	var out Flaw
	path := fmt.Sprintf("/api/v1/client/%s/report/%s/flaw/%s", clientID, reportID, flawID)
	if err := s.c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create POSTs a new flaw into the report.
func (s *FlawService) Create(ctx context.Context, clientID, reportID string, f *Flaw) (*Flaw, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	var out Flaw
	path := fmt.Sprintf("/api/v1/client/%s/report/%s/flaw/create", clientID, reportID)
	if err := s.c.do(ctx, http.MethodPost, path, f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update PUTs changes to an existing flaw.
func (s *FlawService) Update(ctx context.Context, clientID, reportID, flawID string, f *Flaw) (*Flaw, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	if err := validateID(flawID); err != nil {
		return nil, err
	}
	var out Flaw
	path := fmt.Sprintf("/api/v1/client/%s/report/%s/flaw/%s", clientID, reportID, flawID)
	if err := s.c.do(ctx, http.MethodPut, path, f, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete removes a flaw.
func (s *FlawService) Delete(ctx context.Context, clientID, reportID, flawID string) error {
	if err := validateID(clientID); err != nil {
		return err
	}
	if err := validateID(reportID); err != nil {
		return err
	}
	if err := validateID(flawID); err != nil {
		return err
	}
	path := fmt.Sprintf("/api/v1/client/%s/report/%s/flaw/%s", clientID, reportID, flawID)
	return s.c.do(ctx, http.MethodDelete, path, nil, nil)
}

// UpsertResult reports the outcome of a single upsert step.
type UpsertResult struct {
	Index  int
	Flaw   *Flaw
	Action string // "create" or "update"
	Err    error
}

// BatchUpsert creates or updates the given flaws in parallel. The index map
// from externalKey -> existing flaw_id drives the create/update decision.
// Workers defaults to Client.Concurrency when <= 0.
func (s *FlawService) BatchUpsert(ctx context.Context, clientID, reportID string, flaws []*Flaw, index map[string]string, workers int) ([]UpsertResult, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	if workers <= 0 {
		workers = s.c.concurrency
	}
	results := make([]UpsertResult, len(flaws))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for i, f := range flaws {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int, f *Flaw) {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				results[i] = UpsertResult{Index: i, Err: ctx.Err()}
				return
			}
			action := "create"
			key := f.ID
			if key == "" {
				key = f.Title
			}
			var (
				out *Flaw
				err error
			)
			if id, ok := index[key]; ok {
				action = "update"
				out, err = s.Update(ctx, clientID, reportID, id, f)
			} else {
				out, err = s.Create(ctx, clientID, reportID, f)
			}
			results[i] = UpsertResult{Index: i, Flaw: out, Action: action, Err: err}
		}(i, f)
	}
	wg.Wait()
	return results, nil
}
