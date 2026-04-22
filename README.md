# UniRate Go Client

Official Go client for the [UniRate API](https://unirateapi.com) — free, real-time and historical currency exchange rates plus VAT rates.

- Real-time exchange rates between 170+ currencies (fiat + crypto)
- Historical rates back to 1999
- Time-series ranges up to 5 years
- Currency conversion (current and historical)
- VAT rates for countries worldwide
- Free tier, no credit card required
- Idiomatic Go: `context.Context` on every method, sentinel errors, no external dependencies

## Requirements

- Go 1.21 or newer

## Installation

```bash
go get github.com/UniRate-API/unirate-api-go
```

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    unirate "github.com/UniRate-API/unirate-api-go"
)

func main() {
    client := unirate.New(os.Getenv("UNIRATE_API_KEY"))
    ctx := context.Background()

    rate, err := client.GetRate(ctx, "USD", "EUR")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("USD -> EUR: %.4f\n", rate)

    euros, err := client.Convert(ctx, 100, "USD", "EUR")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("100 USD = %.2f EUR\n", euros)
}
```

Get a free API key at [https://unirateapi.com](https://unirateapi.com).

## API

### Construction

```go
client := unirate.New("your-api-key",
    unirate.WithTimeout(10 * time.Second),        // override default 30s
    unirate.WithHTTPClient(customHTTPClient),     // inject your own *http.Client (or any HTTPDoer)
    unirate.WithBaseURL("https://api.example"),   // override base URL (for tests)
)
```

### Current rates

```go
rate, err := client.GetRate(ctx, "USD", "EUR")
// → float64

rates, err := client.GetAllRates(ctx, "USD")
// → map[string]float64

result, err := client.Convert(ctx, 100, "USD", "EUR")
// → float64

codes, err := client.GetSupportedCurrencies(ctx)
// → []string
```

### Historical data (Pro-gated — 403 on free tier)

```go
rate, err := client.GetHistoricalRate(ctx, "2024-01-01", "USD", "EUR")

rates, err := client.GetHistoricalRates(ctx, "2024-01-01", "USD")

amount, err := client.ConvertHistorical(ctx, 100, "USD", "EUR", "2024-01-01")

series, err := client.GetTimeSeries(ctx,
    "2024-01-01", "2024-01-07",
    1,                        // amount
    "USD",                    // base
    []string{"EUR", "GBP"},   // pass nil for all
)

limits, err := client.GetHistoricalLimits(ctx)
```

### VAT rates

```go
all, err := client.GetVATRates(ctx)

de, err := client.GetVATRate(ctx, "DE")
fmt.Println(de.VATData.VATRate) // 19.0
```

### `format` and `callback`

Every method accepts an optional `CallOptions` struct as a trailing
variadic argument:

```go
client.GetRate(ctx, "USD", "EUR",
    unirate.CallOptions{Format: "xml"},
)
```

Non-JSON formats will cause the typed decode to fail — use a custom
`*http.Client` and build the URL yourself if you need raw XML/CSV bodies.

## Error handling

The client returns sentinel errors wrapped with [`fmt.Errorf`], so use
[`errors.Is`] to classify them and [`errors.As`] to extract HTTP details:

```go
_, err := client.GetRate(ctx, "USD", "ZZZ")
switch {
case errors.Is(err, unirate.ErrAuthentication):
    // invalid API key
case errors.Is(err, unirate.ErrInvalidCurrency):
    // unknown currency code
case errors.Is(err, unirate.ErrRateLimit):
    // back off and retry
case errors.Is(err, unirate.ErrInvalidDate):
    // bad date format
default:
    var apiErr *unirate.APIError
    if errors.As(err, &apiErr) {
        // generic HTTP error with apiErr.StatusCode / apiErr.Body,
        // including 403 on Pro-gated endpoints
    }
}
```

## Dependency injection for tests

The `WithHTTPClient` option accepts anything that implements `Do(*http.Request) (*http.Response, error)`, which `*http.Client` already satisfies. Combined with `WithBaseURL`, that makes it trivial to point the client at an `httptest.Server`:

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    _, _ = w.Write([]byte(`{"rate": "0.9"}`))
}))
defer srv.Close()

client := unirate.New("test-key",
    unirate.WithBaseURL(srv.URL),
    unirate.WithHTTPClient(srv.Client()),
)
```

## Rate limits

- **Currency endpoints:** standard rate limits apply
- **Historical endpoints:** 50 requests/hour on the free tier
- **VAT endpoints:** 1800 requests/hour on the free tier

## Testing

```bash
go test ./...                       # mock tests (no network)
UNIRATE_API_KEY=... go test -tags live ./...   # live tests (free-tier endpoints)
```

## Related clients

- [unirate-api-python](https://github.com/UniRate-API/unirate-api-python) (PyPI: `unirate-api`)
- [unirate-api-nodejs](https://github.com/UniRate-API/unirate-api-nodejs) (npm: `unirate-api`)
- [unirate-api-swift](https://github.com/UniRate-API/unirate-api-swift)

## License

MIT — see [LICENSE](LICENSE).
