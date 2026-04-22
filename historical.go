package unirate

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// formatFloat renders a float for query strings without trailing zeros.
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// GetHistoricalRate returns the historical rate for a single from/to pair on
// a given date (YYYY-MM-DD). This endpoint is Pro-gated on the API; free-tier
// keys will receive an [APIError] with StatusCode 403.
func (c *Client) GetHistoricalRate(ctx context.Context, date, from, to string, opts ...CallOptions) (float64, error) {
	q := url.Values{}
	q.Set("date", date)
	q.Set("amount", "1")
	q.Set("from", upper(from))
	q.Set("to", upper(to))
	for _, o := range opts {
		o.apply(q)
	}
	var resp struct {
		Rate flexFloat `json:"rate"`
	}
	if _, err := c.doJSON(ctx, "/api/historical/rates", q, &resp); err != nil {
		return 0, err
	}
	return float64(resp.Rate), nil
}

// GetHistoricalRates returns all rates for a base on a given date.
// Pro-gated.
func (c *Client) GetHistoricalRates(ctx context.Context, date, base string, opts ...CallOptions) (map[string]float64, error) {
	q := url.Values{}
	q.Set("date", date)
	q.Set("amount", "1")
	q.Set("from", upper(base))
	for _, o := range opts {
		o.apply(q)
	}
	var resp struct {
		Rates flexFloatMap `json:"rates"`
	}
	if _, err := c.doJSON(ctx, "/api/historical/rates", q, &resp); err != nil {
		return nil, err
	}
	return resp.Rates.toFloatMap(), nil
}

// ConvertHistorical converts an amount using a rate from a specific date.
// Pro-gated.
func (c *Client) ConvertHistorical(ctx context.Context, amount float64, from, to, date string, opts ...CallOptions) (float64, error) {
	q := url.Values{}
	q.Set("date", date)
	q.Set("amount", formatFloat(amount))
	q.Set("from", upper(from))
	q.Set("to", upper(to))
	for _, o := range opts {
		o.apply(q)
	}
	var resp struct {
		Result flexFloat `json:"result"`
	}
	if _, err := c.doJSON(ctx, "/api/historical/rates", q, &resp); err != nil {
		return 0, err
	}
	return float64(resp.Result), nil
}

// GetTimeSeries fetches a range of historical rates. Pro-gated.
//
// The returned struct mirrors the API shape; most callers only need the
// `Data` field (date -> currency -> rate).
func (c *Client) GetTimeSeries(ctx context.Context, startDate, endDate string, amount float64, base string, currencies []string, opts ...CallOptions) (*TimeSeriesData, error) {
	q := url.Values{}
	q.Set("start_date", startDate)
	q.Set("end_date", endDate)
	q.Set("amount", formatFloat(amount))
	q.Set("base", upper(base))
	if len(currencies) > 0 {
		upper := make([]string, len(currencies))
		for i, cc := range currencies {
			upper[i] = strings.ToUpper(cc)
		}
		q.Set("currencies", strings.Join(upper, ","))
	}
	for _, o := range opts {
		o.apply(q)
	}
	var resp TimeSeriesData
	if _, err := c.doJSON(ctx, "/api/historical/timeseries", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetHistoricalLimits returns per-currency historical coverage. Pro-gated.
func (c *Client) GetHistoricalLimits(ctx context.Context, opts ...CallOptions) (*HistoricalLimits, error) {
	q := url.Values{}
	for _, o := range opts {
		o.apply(q)
	}
	var resp HistoricalLimits
	if _, err := c.doJSON(ctx, "/api/historical/limits", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
