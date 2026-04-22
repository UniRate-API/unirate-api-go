//go:build live

package unirate

import (
	"context"
	"os"
	"testing"
	"time"
)

// Live integration tests hit api.unirateapi.com. Run with:
//
//	UNIRATE_API_KEY=your-key go test -tags live ./...
//
// Each test skips if UNIRATE_API_KEY is empty. Only free-tier endpoints
// (rates, convert, currencies, vat) are exercised — historical and
// timeseries are Pro-gated and would 403 on a free-tier key.

func liveClient(t *testing.T) *Client {
	t.Helper()
	key := os.Getenv("UNIRATE_API_KEY")
	if key == "" {
		t.Skip("UNIRATE_API_KEY not set; skipping live test")
	}
	return New(key, WithTimeout(20*time.Second))
}

func TestLiveGetRate(t *testing.T) {
	client := liveClient(t)
	rate, err := client.GetRate(context.Background(), "USD", "EUR")
	if err != nil {
		t.Fatalf("GetRate: %v", err)
	}
	if rate <= 0 || rate >= 10 {
		t.Errorf("rate out of plausible range: %v", rate)
	}
}

func TestLiveGetAllRates(t *testing.T) {
	client := liveClient(t)
	rates, err := client.GetAllRates(context.Background(), "USD")
	if err != nil {
		t.Fatalf("GetAllRates: %v", err)
	}
	if _, ok := rates["EUR"]; !ok {
		t.Error("expected EUR in all rates")
	}
	if len(rates) < 100 {
		t.Errorf("only %d rates returned; expected 100+", len(rates))
	}
}

func TestLiveConvert(t *testing.T) {
	client := liveClient(t)
	result, err := client.Convert(context.Background(), 100, "USD", "EUR")
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if result <= 0 || result >= 1000 {
		t.Errorf("result out of plausible range: %v", result)
	}
}

func TestLiveSupportedCurrencies(t *testing.T) {
	client := liveClient(t)
	codes, err := client.GetSupportedCurrencies(context.Background())
	if err != nil {
		t.Fatalf("GetSupportedCurrencies: %v", err)
	}
	found := map[string]bool{}
	for _, c := range codes {
		found[c] = true
	}
	if !found["USD"] || !found["EUR"] {
		t.Errorf("expected USD and EUR in supported currencies (got %d total)", len(codes))
	}
}

func TestLiveVATCountry(t *testing.T) {
	client := liveClient(t)
	resp, err := client.GetVATRate(context.Background(), "DE")
	if err != nil {
		t.Fatalf("GetVATRate: %v", err)
	}
	if resp.VATData.CountryCode != "DE" || resp.VATData.CountryName != "Germany" {
		t.Errorf("unexpected response: %+v", resp)
	}
	if resp.VATData.VATRate != 19.0 {
		t.Errorf("VATRate = %v, want 19.0", resp.VATData.VATRate)
	}
}

func TestLiveVATAllCountries(t *testing.T) {
	client := liveClient(t)
	resp, err := client.GetVATRates(context.Background())
	if err != nil {
		t.Fatalf("GetVATRates: %v", err)
	}
	if resp.TotalCountries < 20 {
		t.Errorf("TotalCountries = %d, want 20+", resp.TotalCountries)
	}
	if resp.VATRates["DE"].CountryName != "Germany" {
		t.Errorf("DE missing or wrong: %+v", resp.VATRates["DE"])
	}
}
