package unirate

import (
	"context"
	"net/url"
)

// GetVATRates returns VAT data for every supported country.
func (c *Client) GetVATRates(ctx context.Context, opts ...CallOptions) (*VATRatesResponse, error) {
	q := url.Values{}
	for _, o := range opts {
		o.apply(q)
	}
	var resp VATRatesResponse
	if _, err := c.doJSON(ctx, "/api/vat/rates", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetVATRate returns VAT data for a single country. The code is a
// case-insensitive ISO-3166 alpha-2 string (e.g. "DE").
func (c *Client) GetVATRate(ctx context.Context, country string, opts ...CallOptions) (*VATCountryResponse, error) {
	q := url.Values{}
	q.Set("country", upper(country))
	for _, o := range opts {
		o.apply(q)
	}
	var resp VATCountryResponse
	if _, err := c.doJSON(ctx, "/api/vat/rates", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
