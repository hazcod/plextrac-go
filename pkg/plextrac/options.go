package plextrac

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
)

// WithHTTPClient overrides the default http.Client used for API calls.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) error {
		if h != nil {
			c.httpClient = h
		}
		return nil
	}
}

// WithUserAgent sets the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *Client) error { c.userAgent = ua; return nil }
}

// WithLogger installs a logr.Logger. Verbosity levels: 1 = request log,
// 2 = retry/backoff detail.
func WithLogger(l logr.Logger) Option {
	return func(c *Client) error { c.log = l; return nil }
}

// WithRetry sets the max retry count and the initial backoff used between
// retries of 5xx and 429 responses. Backoff doubles with each attempt.
func WithRetry(retries int, initialBackoff time.Duration) Option {
	return func(c *Client) error {
		if retries < 0 {
			retries = 0
		}
		if initialBackoff <= 0 {
			initialBackoff = defaultBackoff
		}
		c.retries = retries
		c.backoff = initialBackoff
		return nil
	}
}

// WithRateLimit caps requests to N per second.
func WithRateLimit(rps int) Option {
	return func(c *Client) error {
		if rps <= 0 {
			rps = defaultRPS
		}
		c.limiter = rate.NewLimiter(rate.Limit(rps), rps)
		return nil
	}
}

// WithConcurrency sets the default worker count for batch operations.
func WithConcurrency(n int) Option {
	return func(c *Client) error {
		if n < 1 {
			n = 1
		}
		c.concurrency = n
		return nil
	}
}

// WithAllowHTTP permits a non-HTTPS base URL. Intended for testing against
// local fakes. Never use in production.
func WithAllowHTTP() Option {
	return func(c *Client) error { c.allowHTTP = true; return nil }
}

// WithInsecureSkipVerify disables TLS certificate validation. Dangerous;
// retained only for self-signed enterprise tenants. Logs a warning.
func WithInsecureSkipVerify() Option {
	return func(c *Client) error {
		c.insecureTLS = true
		c.log.Info("WARNING: TLS verification disabled")
		return nil
	}
}

// WithDebugBody includes the raw response body in APIError.Body. Off by
// default to avoid leaking sensitive fields through error returns.
func WithDebugBody(enabled bool) Option {
	return func(c *Client) error { c.debugBody = enabled; return nil }
}

func cloneClientWithInsecureTLS(h *http.Client) *http.Client {
	t, ok := h.Transport.(*http.Transport)
	if !ok || t == nil {
		t = http.DefaultTransport.(*http.Transport).Clone()
	} else {
		t = t.Clone()
	}
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	t.TLSClientConfig.InsecureSkipVerify = true
	cp := *h
	cp.Transport = t
	return &cp
}
