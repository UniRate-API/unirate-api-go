package unirate

// Version is the released client version. Exposed so callers and tests can
// assert the User-Agent.
const Version = "0.1.0"

// HistoricalLimit describes the historical data coverage for a single
// currency.
type HistoricalLimit struct {
	EarliestDate string `json:"earliest_date"`
	LatestDate   string `json:"latest_date"`
	TotalDays    int    `json:"total_days,omitempty"`
	Description  string `json:"description,omitempty"`
}

// HistoricalLimits is the response body of GetHistoricalLimits.
type HistoricalLimits struct {
	TotalCurrencies int                        `json:"total_currencies"`
	DataSource      string                     `json:"data_source,omitempty"`
	Currencies      map[string]HistoricalLimit `json:"currencies"`
}

// VATRate describes a single country's VAT rate.
type VATRate struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	VATRate     float64 `json:"vat_rate"`
}

// VATCountryResponse is the response for a per-country VAT lookup.
type VATCountryResponse struct {
	Country string  `json:"country"`
	VATData VATRate `json:"vat_data"`
}

// VATRatesResponse is the response for the all-countries VAT endpoint.
type VATRatesResponse struct {
	Date           string             `json:"date,omitempty"`
	TotalCountries int                `json:"total_countries"`
	VATRates       map[string]VATRate `json:"vat_rates"`
}

// TimeSeriesData is the parsed response of GetTimeSeries. Keys of Data are
// dates in YYYY-MM-DD, inner keys are currency codes.
type TimeSeriesData struct {
	Amount     float64                       `json:"amount"`
	Base       string                        `json:"base"`
	StartDate  string                        `json:"start_date"`
	EndDate    string                        `json:"end_date"`
	TotalDays  int                           `json:"total_days"`
	Currencies []string                      `json:"currencies"`
	Data       map[string]map[string]float64 `json:"data"`
}
