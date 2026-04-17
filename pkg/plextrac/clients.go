package plextrac

import (
	"context"
	"fmt"
	"net/http"
)

// ClientService covers /client* endpoints.
type ClientService struct{ c *Client }

func (s *ClientService) List(ctx context.Context, opts ListOpts) *Iter[ClientOrg] {
	return newIter[ClientOrg](ctx, func(ctx context.Context, page int) ([]ClientOrg, bool, error) {
		o := opts
		if o.Page == 0 {
			o.Page = page
		}
		path := "/api/v1/clients"
		if q := o.query().Encode(); q != "" {
			path += "?" + q
		}
		var resp struct {
			Clients []ClientOrg `json:"clients"`
			Data    []ClientOrg `json:"data"`
			HasMore bool        `json:"has_more"`
		}
		if err := s.c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return nil, false, err
		}
		items := resp.Clients
		if items == nil {
			items = resp.Data
		}
		return items, resp.HasMore && opts.Page == 0, nil
	})
}

func (s *ClientService) Get(ctx context.Context, clientID string) (*ClientOrg, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	var out ClientOrg
	if err := s.c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/client/%s", clientID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ClientService) Create(ctx context.Context, client *ClientOrg) (*ClientOrg, error) {
	var out ClientOrg
	if err := s.c.do(ctx, http.MethodPost, "/api/v1/client/create", client, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ClientService) Update(ctx context.Context, clientID string, client *ClientOrg) (*ClientOrg, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	var out ClientOrg
	if err := s.c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/client/%s", clientID), client, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ClientService) Delete(ctx context.Context, clientID string) error {
	if err := validateID(clientID); err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/client/%s", clientID), nil, nil)
}
