// Package plextrac is a Go client for the Plextrac REST API.
//
// Typical use:
//
//	c, err := plextrac.New("https://tenant.kevlar.plextrac.com",
//	    plextrac.WithPasswordAuth("user", "pass", plextrac.EnvMFA{}),
//	    plextrac.WithRetry(3, time.Second),
//	)
//	if err != nil { return err }
//	flaw, err := c.Flaws.Get(ctx, clientID, reportID, flawID)
package plextrac

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
)

const (
	defaultUserAgent = "plextrac-go/0.0.0"
	defaultTimeout   = 60 * time.Second
	defaultRetries   = 3
	defaultBackoff   = time.Second
	defaultRPS       = 10
)

var idPattern = regexp.MustCompile(`^[0-9a-zA-Z_-]+$`)

// Client is the top-level Plextrac SDK handle. It is safe for concurrent use.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	userAgent  string
	log        logr.Logger

	auth          authProvider
	retries       int
	backoff       time.Duration
	limiter       *rate.Limiter
	concurrency   int
	allowHTTP     bool
	insecureTLS   bool
	debugBody     bool

	// Lazily-initialised resource services.
	Flaws       *FlawService
	Clients     *ClientService
	Reports     *ReportService
	Assets      *AssetService
	Writeups    *WriteupService
	Attachments *AttachmentService
	Templates   *TemplateService
	Tags        *TagService
	Users       *UserService
	Exports     *ExportService
}

// Option configures a Client.
type Option func(*Client) error

// New creates a Client pointing at the given Plextrac tenant base URL.
func New(baseURL string, opts ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidBaseURL, baseURL)
	}
	c := &Client{
		baseURL:     u,
		httpClient:  &http.Client{Timeout: defaultTimeout},
		userAgent:   defaultUserAgent,
		log:         logr.Discard(),
		retries:     defaultRetries,
		backoff:     defaultBackoff,
		limiter:     rate.NewLimiter(rate.Limit(defaultRPS), defaultRPS),
		concurrency: 4,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if u.Scheme != "https" && !c.allowHTTP {
		return nil, fmt.Errorf("%w: must be https (use WithAllowHTTP to override)", ErrInvalidBaseURL)
	}
	if c.insecureTLS {
		c.httpClient = cloneClientWithInsecureTLS(c.httpClient)
	}

	c.Flaws = &FlawService{c: c}
	c.Clients = &ClientService{c: c}
	c.Reports = &ReportService{c: c}
	c.Assets = &AssetService{c: c}
	c.Writeups = &WriteupService{c: c}
	c.Attachments = &AttachmentService{c: c}
	c.Templates = &TemplateService{c: c}
	c.Tags = &TagService{c: c}
	c.Users = &UserService{c: c}
	c.Exports = &ExportService{c: c}
	return c, nil
}

// Concurrency returns the configured worker pool size for batch operations.
func (c *Client) Concurrency() int { return c.concurrency }

// do executes an authenticated, rate-limited, retried HTTP request. The
// response is decoded into out (if non-nil). On non-2xx it returns an
// *APIError. 5xx and 429 responses are retried with exponential backoff.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	endpoint, err := c.resolve(path)
	if err != nil {
		return err
	}
	var bodyBytes []byte
	if body != nil {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return fmt.Errorf("plextrac: encode request: %w", err)
		}
		bodyBytes = buf.Bytes()
	}

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}
		req, err := http.NewRequestWithContext(ctx, method, endpoint, bytesReader(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.userAgent)
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.auth != nil {
			if err := c.auth.apply(ctx, c, req); err != nil {
				return err
			}
		}
		c.log.V(1).Info("http", "method", method, "url", endpoint, "attempt", attempt)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if !c.retryable(0) {
				return err
			}
			c.sleep(ctx, attempt, 0)
			continue
		}
		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out != nil && len(respBody) > 0 {
				if err := json.Unmarshal(respBody, out); err != nil {
					return fmt.Errorf("plextrac: decode response: %w", err)
				}
			}
			return nil
		}

		apiErr := &APIError{StatusCode: resp.StatusCode}
		var parsed struct {
			Message string `json:"message"`
			Error   string `json:"error"`
			Code    string `json:"code"`
		}
		if json.Unmarshal(respBody, &parsed) == nil {
			apiErr.Code = parsed.Code
			apiErr.Message = firstNonEmpty(parsed.Message, parsed.Error)
		}
		if c.debugBody {
			apiErr.Body = respBody
		}
		lastErr = apiErr

		if !c.retryable(resp.StatusCode) || attempt == c.retries {
			return apiErr
		}
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		c.sleep(ctx, attempt, retryAfter)
	}
	return lastErr
}

func (c *Client) retryable(status int) bool {
	if status == 0 {
		return true
	}
	if status == http.StatusTooManyRequests {
		return true
	}
	return status >= 500
}

func (c *Client) sleep(ctx context.Context, attempt int, override time.Duration) {
	d := override
	if d <= 0 {
		d = c.backoff << attempt
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

func (c *Client) resolve(path string) (string, error) {
	p, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	return c.baseURL.ResolveReference(p).String(), nil
}

// validateID rejects IDs containing characters that could alter URL structure.
func validateID(id string) error {
	if id == "" || !idPattern.MatchString(id) {
		return fmt.Errorf("%w: %q", ErrInvalidID, id)
	}
	return nil
}

func bytesReader(b []byte) io.Reader {
	if b == nil {
		return nil
	}
	return bytes.NewReader(b)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 0
	}
	// seconds form only; HTTP-date form is ignored
	var secs int
	if _, err := fmt.Sscanf(strings.TrimSpace(h), "%d", &secs); err != nil {
		return 0
	}
	if secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}
