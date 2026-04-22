package unirate

import (
	"context"
	"net/url"
)

// GetRate returns the current exchange rate between `from` and `to`.
//
// Codes are uppercased before being sent.
func (c *Client) GetRate(ctx context.Context, from, to string, opts ...CallOptions) (float64, error) {
	q := url.Values{}
	q.Set("from", upper(from))
	q.Set("to", upper(to))
	for _, o := range opts {
		o.apply(q)
	}
	var resp struct {
		Rate flexFloat `json:"rate"`
	}
	if _, err := c.doJSON(ctx, "/api/rates", q, &resp); err != nil {
		return 0, err
	}
	return float64(resp.Rate), nil
}

// GetAllRates returns all rates for the given base currency.
func (c *Client) GetAllRates(ctx context.Context, from string, opts ...CallOptions) (map[string]float64, error) {
	q := url.Values{}
	q.Set("from", upper(from))
	for _, o := range opts {
		o.apply(q)
	}
	var resp struct {
		Rates flexFloatMap `json:"rates"`
	}
	if _, err := c.doJSON(ctx, "/api/rates", q, &resp); err != nil {
		return nil, err
	}
	return resp.Rates.toFloatMap(), nil
}

// Convert converts an amount from one currency to another using the current
// rate.
func (c *Client) Convert(ctx context.Context, amount float64, from, to string, opts ...CallOptions) (float64, error) {
	q := url.Values{}
	q.Set("from", upper(from))
	q.Set("to", upper(to))
	q.Set("amount", formatFloat(amount))
	for _, o := range opts {
		o.apply(q)
	}
	var resp struct {
		Result flexFloat `json:"result"`
	}
	if _, err := c.doJSON(ctx, "/api/convert", q, &resp); err != nil {
		return 0, err
	}
	return float64(resp.Result), nil
}

// GetSupportedCurrencies returns the list of currency codes the API knows
// about.
func (c *Client) GetSupportedCurrencies(ctx context.Context, opts ...CallOptions) ([]string, error) {
	q := url.Values{}
	for _, o := range opts {
		o.apply(q)
	}
	var resp struct {
		Currencies []string `json:"currencies"`
	}
	if _, err := c.doJSON(ctx, "/api/currencies", q, &resp); err != nil {
		return nil, err
	}
	return resp.Currencies, nil
}
