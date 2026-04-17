package plextrac

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

// authProvider applies credentials to an outgoing request. Implementations
// must be safe for concurrent use; token refresh happens under a lock.
type authProvider interface {
	apply(ctx context.Context, c *Client, req *http.Request) error
}

// WithToken supplies a pre-obtained JWT.
func WithToken(token string) Option {
	return func(c *Client) error {
		if token == "" {
			return ErrNoCredentials
		}
		c.auth = &staticToken{token: token}
		return nil
	}
}

// WithAPIKey authenticates with a long-lived API key header (equivalent to
// a bearer token on this tenant).
func WithAPIKey(key string) Option {
	return func(c *Client) error {
		if key == "" {
			return ErrNoCredentials
		}
		c.auth = &staticToken{token: key}
		return nil
	}
}

// WithPasswordAuth performs an interactive-style password login, exchanging
// the password for a token on first use. The password is kept only in the
// closure of the MFA exchange and discarded once a token is obtained.
func WithPasswordAuth(username, password string, mfa MFAProvider) Option {
	return func(c *Client) error {
		if username == "" || password == "" {
			return ErrNoCredentials
		}
		c.auth = &passwordAuth{
			user: username,
			pass: password,
			mfa:  mfa,
		}
		return nil
	}
}

// MFAProvider supplies an MFA/TOTP code in response to a challenge string.
type MFAProvider interface {
	Code(challenge string) (string, error)
}

// EnvMFA reads the code from PLEXTRAC_MFA at challenge time.
type EnvMFA struct{ Var string }

func (e EnvMFA) Code(_ string) (string, error) {
	name := e.Var
	if name == "" {
		name = "PLEXTRAC_MFA"
	}
	code := os.Getenv(name)
	if code == "" {
		return "", ErrMFARequired
	}
	return code, nil
}

// StaticMFA returns a preset code.
type StaticMFA string

func (s StaticMFA) Code(_ string) (string, error) {
	if s == "" {
		return "", ErrMFARequired
	}
	return string(s), nil
}

type staticToken struct{ token string }

func (s *staticToken) apply(_ context.Context, _ *Client, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+s.token)
	return nil
}

type passwordAuth struct {
	mu        sync.Mutex
	user      string
	pass      string
	mfa       MFAProvider
	token     string
	exhausted bool
}

func (p *passwordAuth) apply(ctx context.Context, c *Client, req *http.Request) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.token == "" {
		tok, err := p.login(ctx, c)
		if err != nil {
			return err
		}
		p.token = tok
		// Minimise lifetime of the raw password in memory. We still need
		// it for re-auth on 401, so zero the backing array only on
		// successful first login and mark as exhausted.
		zero(&p.pass)
		p.exhausted = true
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	return nil
}

func (p *passwordAuth) login(ctx context.Context, c *Client) (string, error) {
	if p.exhausted {
		return "", fmt.Errorf("plextrac: password not retained; re-create Client to re-auth")
	}
	body := map[string]string{"username": p.user, "password": p.pass}
	var resp struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
		MFAEnabled  bool   `json:"mfa_enabled"`
		Code        string `json:"code"`
		MFAToken    string `json:"mfa_token"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/v1/authenticate", body, &resp); err != nil {
		return "", err
	}
	if resp.Token != "" && !resp.MFAEnabled {
		return resp.Token, nil
	}
	if resp.AccessToken != "" && !resp.MFAEnabled {
		return resp.AccessToken, nil
	}
	challenge := resp.Code
	if challenge == "" {
		challenge = resp.MFAToken
	}
	if challenge == "" {
		return "", fmt.Errorf("plextrac: login returned neither token nor MFA challenge")
	}
	if p.mfa == nil {
		return "", ErrMFARequired
	}
	code, err := p.mfa.Code(challenge)
	if err != nil {
		return "", err
	}
	mfaBody := map[string]string{"code": challenge, "token": code}
	var mfaResp struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/v1/authenticate/mfa", mfaBody, &mfaResp); err != nil {
		return "", err
	}
	tok := mfaResp.Token
	if tok == "" {
		tok = mfaResp.AccessToken
	}
	if tok == "" {
		return "", fmt.Errorf("plextrac: MFA exchange returned no token")
	}
	return tok, nil
}

// zero overwrites the backing bytes of a string holder before release.
// Go strings are immutable so this operates on the captured password via
// a pointer; it is best-effort — the GC may have already copied the value.
func zero(s *string) {
	b := []byte(*s)
	for i := range b {
		b[i] = 0
	}
	*s = strings.Repeat("\x00", len(b))
}
