package plextrac

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// AttachmentService covers file upload/download endpoints.
type AttachmentService struct{ c *Client }

// Upload posts a file to a flaw as an attachment. The reader is consumed
// once; the caller is responsible for seeking/resetting if retry is desired.
func (s *AttachmentService) Upload(ctx context.Context, clientID, reportID, flawID, filename string, body io.Reader) (*Attachment, error) {
	if err := validateID(clientID); err != nil {
		return nil, err
	}
	if err := validateID(reportID); err != nil {
		return nil, err
	}
	if err := validateID(flawID); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, body); err != nil {
		return nil, err
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	endpoint, err := s.c.resolve(fmt.Sprintf("/api/v1/client/%s/report/%s/flaw/%s/upload", clientID, reportID, flawID))
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", s.c.userAgent)
	if s.c.auth != nil {
		if err := s.c.auth.apply(ctx, s.c, req); err != nil {
			return nil, err
		}
	}
	if err := s.c.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	resp, err := s.c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if s.c.debugBody {
			apiErr.Body = respBody
		}
		return nil, apiErr
	}
	var out Attachment
	if len(respBody) > 0 {
		if err := jsonUnmarshal(respBody, &out); err != nil {
			return nil, err
		}
	}
	return &out, nil
}

// Download streams an attachment to the given writer.
func (s *AttachmentService) Download(ctx context.Context, clientID, reportID, attachmentID string, w io.Writer) error {
	if err := validateID(clientID); err != nil {
		return err
	}
	if err := validateID(reportID); err != nil {
		return err
	}
	if err := validateID(attachmentID); err != nil {
		return err
	}
	endpoint, err := s.c.resolve(fmt.Sprintf("/api/v1/client/%s/report/%s/attachment/%s", clientID, reportID, attachmentID))
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", s.c.userAgent)
	if s.c.auth != nil {
		if err := s.c.auth.apply(ctx, s.c, req); err != nil {
			return err
		}
	}
	if err := s.c.limiter.Wait(ctx); err != nil {
		return err
	}
	resp, err := s.c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return &APIError{StatusCode: resp.StatusCode}
	}
	_, err = io.Copy(w, resp.Body)
	return err
}
