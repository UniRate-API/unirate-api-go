// Package unirate is the official Go client for the UniRate API
// (https://unirateapi.com) — free currency exchange rates, historical data,
// and VAT rates.
//
// Create a client with [New] and call the typed methods:
//
//	client := unirate.New("your-api-key")
//	rate, err := client.GetRate(ctx, "USD", "EUR")
//
// All methods take a [context.Context] as the first parameter, propagate
// cancellation to the underlying HTTP call, and return sentinel errors
// (see errors.go) wrapped with the relevant detail. Use [errors.Is] /
// [errors.As] to classify failures.
package unirate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// DefaultBaseURL is the production UniRate API endpoint.
const DefaultBaseURL = "https://api.unirateapi.com"

// DefaultTimeout is the default per-request timeout when no custom
// [http.Client] is supplied.
const DefaultTimeout = 30 * time.Second

// HTTPDoer is the minimal interface the Client needs from an HTTP client.
// *http.Client satisfies this, which is all that matters in practice — it
// exists chiefly so tests can inject an httptest server's client without
// gymnastics.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client is a UniRate API client. Create one with [New]; it is safe for
// concurrent use by multiple goroutines.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient HTTPDoer
	userAgent  string
}

// Option configures a Client during construction.
type Option func(*Client)

// WithTimeout sets the request timeout. Ignored if [WithHTTPClient] is also
// provided (in that case the caller controls timeouts on their own client).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if hc, ok := c.httpClient.(*http.Client); ok {
			hc.Timeout = d
		}
	}
}

// WithHTTPClient replaces the default *http.Client. Useful for tests
// (inject httptest.NewServer().Client()) and for callers who want custom
// transports, proxies, or timeouts.
func WithHTTPClient(h HTTPDoer) Option {
	return func(c *Client) {
		c.httpClient = h
	}
}

// WithBaseURL overrides the API base URL. Mostly useful for tests against
// an httptest server.
func WithBaseURL(u string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(u, "/")
	}
}

// New constructs a Client with the given API key and applies any options.
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		userAgent: "unirate-go/" + Version,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// CallOptions are per-call extras for `format` and `callback` query
// parameters defined by the API. When Format is non-empty and not "json",
// the response body is returned as a raw string via the `*Raw` sibling
// methods; the typed methods will fall back to raw passthrough only if
// Format is unset or explicitly "json".
//
// Most callers can ignore this and use the default JSON paths.
type CallOptions struct {
	Format   string
	Callback string
}

// callOption is a functional option type for per-call extras — kept private
// so the public surface stays consistent with the CallOptions struct form.
func (o CallOptions) apply(q url.Values) {
	if o.Format != "" {
		q.Set("format", o.Format)
	}
	if o.Callback != "" {
		q.Set("callback", o.Callback)
	}
}

// flexFloat decodes either a JSON number or a JSON string containing a
// number. The API returns rate values as strings in some places (e.g.
// /api/rates) and as floats in others — this keeps decoding consistent.
type flexFloat float64

func (f *flexFloat) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	// String form: "0.92"
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		if s == "" {
			*f = 0
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("unirate: invalid numeric string %q: %w", s, err)
		}
		*f = flexFloat(v)
		return nil
	}
	// Number form: 0.92
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*f = flexFloat(v)
	return nil
}

// flexFloatMap maps string keys to flexFloat.
type flexFloatMap map[string]flexFloat

func (m flexFloatMap) toFloatMap() map[string]float64 {
	out := make(map[string]float64, len(m))
	for k, v := range m {
		out[k] = float64(v)
	}
	return out
}

// buildURL assembles the fully-qualified URL for a path + query map. The
// api_key query parameter is always appended.
func (c *Client) buildURL(path string, q url.Values) (string, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("unirate: invalid base URL: %w", err)
	}
	if q == nil {
		q = url.Values{}
	}
	q.Set("api_key", c.apiKey)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// doJSON performs a GET request and decodes the response body into out.
// out may be nil to signal "return the raw body as a string instead".
func (c *Client) doJSON(ctx context.Context, path string, q url.Values, out any) ([]byte, error) {
	fullURL, err := c.buildURL(path, q)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unirate: building request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unirate: network: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unirate: reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, c.mapStatus(resp.StatusCode, body)
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return body, fmt.Errorf("unirate: decoding response: %w", err)
		}
	}
	return body, nil
}

// mapStatus converts an HTTP status code into the appropriate error. It
// always returns a non-nil error. For the documented status codes it
// wraps the sentinel so callers can use errors.Is; it also attaches an
// *APIError so errors.As can retrieve StatusCode / Body.
func (c *Client) mapStatus(status int, body []byte) error {
	apiErr := &APIError{StatusCode: status, Body: strings.TrimSpace(string(body))}
	switch status {
	case http.StatusBadRequest: // 400
		return fmt.Errorf("%w: %w", ErrInvalidDate, apiErr)
	case http.StatusUnauthorized: // 401
		return fmt.Errorf("%w: %w", ErrAuthentication, apiErr)
	case http.StatusNotFound: // 404
		return fmt.Errorf("%w: %w", ErrInvalidCurrency, apiErr)
	case http.StatusTooManyRequests: // 429
		return fmt.Errorf("%w: %w", ErrRateLimit, apiErr)
	default:
		// 403 Pro-gate, 503, and every other non-2xx land here as a
		// bare APIError so callers can still pull StatusCode/Body via
		// errors.As. We don't wrap with a sentinel because there isn't
		// a more specific one.
		return apiErr
	}
}

// upper is a small helper that uppercases in-place — the API expects upper
// case codes and the existing clients all normalise input.
func upper(s string) string { return strings.ToUpper(s) }
