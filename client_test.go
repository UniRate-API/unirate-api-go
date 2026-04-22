package unirate

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestServer spins up an httptest server that runs the given handler,
// plus a Client pointed at it. Callers close the returned server.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client := New(
		"test-key",
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
	)
	return srv, client
}

// assertCommonHeaders checks every request carries the headers we promise.
func assertCommonHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("Accept"); got != "application/json" {
		t.Errorf("Accept header = %q, want application/json", got)
	}
	if got := r.Header.Get("User-Agent"); !strings.HasPrefix(got, "unirate-go/") {
		t.Errorf("User-Agent = %q, want prefix unirate-go/", got)
	}
	if got := r.URL.Query().Get("api_key"); got != "test-key" {
		t.Errorf("api_key query = %q, want test-key", got)
	}
}

func TestGetRate(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertCommonHeaders(t, r)
		if r.URL.Path != "/api/rates" {
			t.Errorf("path = %q, want /api/rates", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("from") != "USD" || q.Get("to") != "EUR" {
			t.Errorf("query = %v, want from=USD to=EUR", q)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"rate": "0.9321"}`))
	})
	defer srv.Close()

	rate, err := client.GetRate(context.Background(), "usd", "eur")
	if err != nil {
		t.Fatalf("GetRate error: %v", err)
	}
	if rate < 0.9320 || rate > 0.9322 {
		t.Errorf("rate = %v, want ~0.9321", rate)
	}
}

func TestGetAllRates(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rates": {"EUR": "0.9", "GBP": 0.8}}`))
	})
	defer srv.Close()

	rates, err := client.GetAllRates(context.Background(), "USD")
	if err != nil {
		t.Fatalf("GetAllRates error: %v", err)
	}
	if rates["EUR"] != 0.9 {
		t.Errorf("EUR = %v, want 0.9", rates["EUR"])
	}
	if rates["GBP"] != 0.8 {
		t.Errorf("GBP = %v, want 0.8", rates["GBP"])
	}
}

func TestConvert(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("amount"); got != "100" {
			t.Errorf("amount = %q, want 100", got)
		}
		_, _ = w.Write([]byte(`{"result": "93.21"}`))
	})
	defer srv.Close()

	v, err := client.Convert(context.Background(), 100, "USD", "EUR")
	if err != nil {
		t.Fatalf("Convert error: %v", err)
	}
	if v < 93.20 || v > 93.22 {
		t.Errorf("result = %v, want ~93.21", v)
	}
}

func TestGetSupportedCurrencies(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assertCommonHeaders(t, r)
		_, _ = w.Write([]byte(`{"currencies": ["USD", "EUR", "GBP", "BTC"]}`))
	})
	defer srv.Close()

	codes, err := client.GetSupportedCurrencies(context.Background())
	if err != nil {
		t.Fatalf("GetSupportedCurrencies error: %v", err)
	}
	if len(codes) != 4 || codes[0] != "USD" {
		t.Errorf("codes = %v", codes)
	}
}

func TestGetHistoricalRate(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("date") != "2024-01-01" {
			t.Errorf("date = %q", q.Get("date"))
		}
		if q.Get("amount") != "1" {
			t.Errorf("amount = %q", q.Get("amount"))
		}
		_, _ = w.Write([]byte(`{"rate": "0.8412"}`))
	})
	defer srv.Close()

	rate, err := client.GetHistoricalRate(context.Background(), "2024-01-01", "USD", "EUR")
	if err != nil {
		t.Fatalf("GetHistoricalRate error: %v", err)
	}
	if rate < 0.841 || rate > 0.842 {
		t.Errorf("rate = %v, want ~0.8412", rate)
	}
}

func TestGetHistoricalRates(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"rates": {"EUR": "0.84", "GBP": "0.77"}}`))
	})
	defer srv.Close()

	rates, err := client.GetHistoricalRates(context.Background(), "2024-01-01", "USD")
	if err != nil {
		t.Fatalf("GetHistoricalRates error: %v", err)
	}
	if rates["EUR"] != 0.84 || rates["GBP"] != 0.77 {
		t.Errorf("rates = %v", rates)
	}
}

func TestConvertHistorical(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result": "84.12"}`))
	})
	defer srv.Close()

	v, err := client.ConvertHistorical(context.Background(), 100, "USD", "EUR", "2024-01-01")
	if err != nil {
		t.Fatalf("ConvertHistorical error: %v", err)
	}
	if v < 84.11 || v > 84.13 {
		t.Errorf("result = %v, want ~84.12", v)
	}
}

func TestGetTimeSeries(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("start_date") != "2024-01-01" || q.Get("end_date") != "2024-01-02" {
			t.Errorf("dates = %v", q)
		}
		if q.Get("currencies") != "EUR,GBP" {
			t.Errorf("currencies = %q, want EUR,GBP", q.Get("currencies"))
		}
		_, _ = w.Write([]byte(`{
			"amount": 1,
			"base": "USD",
			"start_date": "2024-01-01",
			"end_date": "2024-01-02",
			"total_days": 2,
			"currencies": ["EUR","GBP"],
			"data": {"2024-01-01": {"EUR": 0.90, "GBP": 0.79}, "2024-01-02": {"EUR": 0.91, "GBP": 0.80}}
		}`))
	})
	defer srv.Close()

	ts, err := client.GetTimeSeries(context.Background(), "2024-01-01", "2024-01-02", 1, "USD", []string{"eur", "gbp"})
	if err != nil {
		t.Fatalf("GetTimeSeries error: %v", err)
	}
	if ts.Data["2024-01-01"]["EUR"] != 0.90 {
		t.Errorf("2024-01-01 EUR = %v", ts.Data["2024-01-01"]["EUR"])
	}
	if ts.Data["2024-01-02"]["GBP"] != 0.80 {
		t.Errorf("2024-01-02 GBP = %v", ts.Data["2024-01-02"]["GBP"])
	}
}

func TestGetHistoricalLimits(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"total_currencies": 2,
			"currencies": {
				"USD": {"earliest_date": "1999-01-01", "latest_date": "2026-04-20"},
				"EUR": {"earliest_date": "1999-01-01", "latest_date": "2026-04-20"}
			}
		}`))
	})
	defer srv.Close()

	limits, err := client.GetHistoricalLimits(context.Background())
	if err != nil {
		t.Fatalf("GetHistoricalLimits error: %v", err)
	}
	if limits.TotalCurrencies != 2 {
		t.Errorf("total = %d", limits.TotalCurrencies)
	}
	if limits.Currencies["USD"].EarliestDate != "1999-01-01" {
		t.Errorf("USD earliest = %q", limits.Currencies["USD"].EarliestDate)
	}
}

func TestGetVATRate(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("country"); got != "DE" {
			t.Errorf("country = %q, want DE", got)
		}
		_, _ = w.Write([]byte(`{"country": "DE", "vat_data": {"country_code": "DE", "country_name": "Germany", "vat_rate": 19.0}}`))
	})
	defer srv.Close()

	resp, err := client.GetVATRate(context.Background(), "de")
	if err != nil {
		t.Fatalf("GetVATRate error: %v", err)
	}
	if resp.Country != "DE" || resp.VATData.CountryName != "Germany" || resp.VATData.VATRate != 19.0 {
		t.Errorf("resp = %+v", resp)
	}
}

func TestGetVATRates(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"date": "2026-01-22",
			"total_countries": 2,
			"vat_rates": {
				"DE": {"country_code": "DE", "country_name": "Germany", "vat_rate": 19.0},
				"FR": {"country_code": "FR", "country_name": "France",  "vat_rate": 20.0}
			}
		}`))
	})
	defer srv.Close()

	resp, err := client.GetVATRates(context.Background())
	if err != nil {
		t.Fatalf("GetVATRates error: %v", err)
	}
	if resp.TotalCountries != 2 || resp.Date != "2026-01-22" {
		t.Errorf("resp = %+v", resp)
	}
	if resp.VATRates["FR"].CountryName != "France" {
		t.Errorf("FR = %+v", resp.VATRates["FR"])
	}
}

func TestHistoricalPaywall403(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error": "Historical data access requires a Pro subscription"}`))
	})
	defer srv.Close()

	_, err := client.GetHistoricalRate(context.Background(), "2024-01-01", "USD", "EUR")
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError via errors.As, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 403 {
		t.Errorf("StatusCode = %d, want 403", apiErr.StatusCode)
	}
}

func TestAuthenticationError(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	defer srv.Close()

	_, err := client.GetRate(context.Background(), "USD", "EUR")
	if !errors.Is(err, ErrAuthentication) {
		t.Fatalf("expected ErrAuthentication, got %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 401 {
		t.Errorf("APIError not retrievable via errors.As or wrong status: %v", err)
	}
}

func TestRateLimitError(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	defer srv.Close()

	_, err := client.GetRate(context.Background(), "USD", "EUR")
	if !errors.Is(err, ErrRateLimit) {
		t.Fatalf("expected ErrRateLimit, got %v", err)
	}
}

func TestInvalidCurrencyError(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer srv.Close()

	_, err := client.GetRate(context.Background(), "USD", "ZZZ")
	if !errors.Is(err, ErrInvalidCurrency) {
		t.Fatalf("expected ErrInvalidCurrency, got %v", err)
	}
}

func TestInvalidDateError(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	defer srv.Close()

	_, err := client.GetHistoricalRate(context.Background(), "not-a-date", "USD", "EUR")
	if !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}

func TestContextCancellation(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(2 * time.Second):
			_, _ = w.Write([]byte(`{"rate": "1.0"}`))
		}
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := client.GetRate(ctx, "USD", "EUR")
	if err == nil {
		t.Fatal("expected context-cancellation error, got nil")
	}
}
