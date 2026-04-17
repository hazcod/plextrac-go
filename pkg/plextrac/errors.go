package plextrac

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError is returned when the Plextrac API responds with a non-2xx status.
// The raw Body is only populated when WithDebugBody(true) was set, to avoid
// accidentally surfacing sensitive fields through error messages.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	Body       []byte
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("plextrac: HTTP %d: %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("plextrac: HTTP %d", e.StatusCode)
}

// Is allows errors.Is(err, plextrac.ErrNotFound) etc.
func (e *APIError) Is(target error) bool {
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	return t.StatusCode == e.StatusCode
}

var (
	ErrBadRequest   = &APIError{StatusCode: http.StatusBadRequest}
	ErrUnauthorized = &APIError{StatusCode: http.StatusUnauthorized}
	ErrForbidden    = &APIError{StatusCode: http.StatusForbidden}
	ErrNotFound     = &APIError{StatusCode: http.StatusNotFound}
	ErrConflict     = &APIError{StatusCode: http.StatusConflict}
	ErrRateLimited  = &APIError{StatusCode: http.StatusTooManyRequests}

	ErrInvalidBaseURL = errors.New("plextrac: invalid base URL")
	ErrInvalidID      = errors.New("plextrac: invalid resource ID")
	ErrNoCredentials  = errors.New("plextrac: no credentials configured")
	ErrMFARequired    = errors.New("plextrac: MFA code required")
)
