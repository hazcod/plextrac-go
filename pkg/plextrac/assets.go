package plextrac

import (
	"context"
	"fmt"
	"net/http"
)

// AssetService covers /client/{clientID}/asset* endpoints.
type AssetService struct{ c *Client }

func (s *AssetService) List(ctx context.Context, clientID string, opts ListOpts) *Iter[Asset] {
	return newIter[Asset](ctx, func(ctx context.Context, page int) ([]Asset, bool, error) {
		if err := validateID(clientID); err != nil {
			return nil, false, err
		}
		o := opts
		if o.Page == 0 {
			o.Page = page
		}
		path := fmt.Sprintf("/api/v1/client/%s/assets", clientID)
		if q := o.query().Encode(); q != "" {
			path += "?" + q
		}
		var resp struct {
			Assets  []Asset `json:"assets"`
			Data    []Asset `json:"data"`
			HasMore bool    `json:"has_more"`
		}
		if err := s.c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return nil, false, err
		}
		items := resp.Assets
		if items == nil {
			items = resp.Data
		}
		return items, resp.HasMore && opts.Page == 0, nil
	})
}

func (s *AssetService) Get(ctx context.Context, clientID, assetID string) (*Asset, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(assetID); err != nil {
		return nil, err
	}
	var out Asset
	if err := s.c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/client/%s/asset/%s", clientID, assetID), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AssetService) Create(ctx context.Context, clientID string, a *Asset) (*Asset, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	var out Asset
	if err := s.c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/client/%s/asset/create", clientID), a, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AssetService) Update(ctx context.Context, clientID, assetID string, a *Asset) (*Asset, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(assetID); err != nil {
		return nil, err
	}
	var out Asset
	if err := s.c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/client/%s/asset/%s", clientID, assetID), a, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *AssetService) Delete(ctx context.Context, clientID, assetID string) error {
	if err := validateID(clientID); err != nil {
		return err
	}
	if err := validateID(assetID); err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/client/%s/asset/%s", clientID, assetID), nil, nil)
}

// AttachToReport links an asset to a report.
func (s *AssetService) AttachToReport(ctx context.Context, clientID, reportID, assetID string) error {
	if err := validateID(clientID); err != nil {
		return err
	}
	if err := validateID(reportID); err != nil {
		return err
	}
	if err := validateID(assetID); err != nil {
		return err
	}
	return s.c.do(ctx, http.MethodPut, fmt.Sprintf("/api/v1/client/%s/report/%s/asset/%s", clientID, reportID, assetID), nil, nil)
}
