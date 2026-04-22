package unirate

import (
	"errors"
	"fmt"
)

// Sentinel errors. Wrap these with fmt.Errorf("...: %w", Err...) so that
// callers can classify failures with errors.Is.
var (
	// ErrAuthentication is returned when the API rejects the key (HTTP 401).
	ErrAuthentication = errors.New("unirate: authentication failed")

	// ErrRateLimit is returned when the account has exceeded its quota (HTTP 429).
	ErrRateLimit = errors.New("unirate: rate limit exceeded")

	// ErrInvalidCurrency is returned for unknown currency codes or missing
	// data (HTTP 404).
	ErrInvalidCurrency = errors.New("unirate: invalid currency")

	// ErrInvalidDate is returned for malformed date parameters (HTTP 400).
	ErrInvalidDate = errors.New("unirate: invalid date or request parameters")
)

// APIError is returned for non-success HTTP responses that don't map to one
// of the sentinel errors above (e.g. 403 Pro-gated endpoints, 5xx, etc.).
//
// It is also wrapped around the sentinel errors so callers can access the
// status code and body via errors.As.
type APIError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("unirate: API error (status %d)", e.StatusCode)
	}
	return fmt.Sprintf("unirate: API error (status %d): %s", e.StatusCode, e.Body)
}
