// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	"net/http"
	"net/http/httptest"
	"testing"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketCollector_Collect_Good_HistoricalWithFromDate_Good(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			data := coinData{
				ID:     "lethean",
				Symbol: "lthn",
				Name:   "Lethean",
				MarketData: marketData{
					CurrentPrice: map[string]float64{"usd": 0.001},
				},
			}
			_ = json.NewEncoder(w).Encode(data)
		} else {
			// Historical data with FromDate param.
			assert.Contains(t, r.URL.RawQuery, "days=")
			data := historicalData{
				Prices:       [][]float64{{1705305600000, 0.001}},
				MarketCaps:   [][]float64{{1705305600000, 10000}},
				TotalVolumes: [][]float64{{1705305600000, 500}},
			}
			_ = json.NewEncoder(w).Encode(data)
		}
	}))
	defer srv.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = srv.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	mc := &MarketCollector{CoinID: "lethean", Historical: true, FromDate: "2025-01-01"}
	result, err := mc.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 3, result.Items)
}

func TestMarketCollector_Collect_Good_HistoricalInvalidDate_Good(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			data := coinData{
				ID:     "test",
				Symbol: "tst",
				Name:   "Test",
				MarketData: marketData{
					CurrentPrice: map[string]float64{"usd": 1.0},
				},
			}
			_ = json.NewEncoder(w).Encode(data)
		} else {
			// Should fall back to 365 days with invalid date.
			assert.Contains(t, r.URL.RawQuery, "days=365")
			data := historicalData{
				Prices: [][]float64{{1705305600000, 1.0}},
			}
			_ = json.NewEncoder(w).Encode(data)
		}
	}))
	defer srv.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = srv.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	mc := &MarketCollector{CoinID: "test", Historical: true, FromDate: "not-a-date"}
	result, err := mc.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 3, result.Items)
}

func TestMarketCollector_Collect_Bad_HistoricalServerError_Good(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			data := coinData{
				ID:     "test",
				Symbol: "tst",
				Name:   "Test",
				MarketData: marketData{
					CurrentPrice: map[string]float64{"usd": 1.0},
				},
			}
			_ = json.NewEncoder(w).Encode(data)
		} else {
			// Historical endpoint fails.
			w.WriteHeader(http.StatusTooManyRequests)
		}
	}))
	defer srv.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = srv.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	mc := &MarketCollector{CoinID: "test", Historical: true}
	result, err := mc.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)  // current.json + summary.md
	assert.Equal(t, 1, result.Errors) // historical failed
}

func TestMarketCollector_Collect_Good_EmitsEvents_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := coinData{
			ID:     "bitcoin",
			Symbol: "btc",
			Name:   "Bitcoin",
			MarketData: marketData{
				CurrentPrice: map[string]float64{"usd": 50000},
			},
		}
		_ = json.NewEncoder(w).Encode(data)
	}))
	defer srv.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = srv.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	var starts, completes int
	cfg.Dispatcher.On(EventStart, func(e Event) { starts++ })
	cfg.Dispatcher.On(EventComplete, func(e Event) { completes++ })

	mc := &MarketCollector{CoinID: "bitcoin"}
	_, err := mc.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 1, starts)
	assert.Equal(t, 1, completes)
}

func TestMarketCollector_Collect_Good_CancelledContext_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = srv.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mc := &MarketCollector{CoinID: "bitcoin"}
	result, err := mc.Collect(ctx, cfg)

	// Context cancellation causes error in fetchJSON.
	require.NoError(t, err) // outer Collect doesn't return errors from currentData fetch
	assert.Equal(t, 1, result.Errors)
}

func TestFormatMarketSummary_Good_AllFields_Good(t *testing.T) {
	data := &coinData{
		Name:   "Lethean",
		Symbol: "lthn",
		MarketData: marketData{
			CurrentPrice:      map[string]float64{"usd": 0.001},
			MarketCap:         map[string]float64{"usd": 100000},
			TotalVolume:       map[string]float64{"usd": 5000},
			High24h:           map[string]float64{"usd": 0.0015},
			Low24h:            map[string]float64{"usd": 0.0005},
			PriceChange24h:    0.0002,
			PriceChangePct24h: 5.5,
			MarketCapRank:     500,
			CirculatingSupply: 1000000000,
			TotalSupply:       2000000000,
			LastUpdated:       "2025-01-15T12:00:00Z",
		},
	}

	summary := FormatMarketSummary(data)

	assert.Contains(t, summary, "# Lethean (LTHN)")
	assert.Contains(t, summary, "24h Volume")
	assert.Contains(t, summary, "24h High")
	assert.Contains(t, summary, "24h Low")
	assert.Contains(t, summary, "24h Price Change")
	assert.Contains(t, summary, "#500")
	assert.Contains(t, summary, "Circulating Supply")
	assert.Contains(t, summary, "Total Supply")
	assert.Contains(t, summary, "Last updated")
}

func TestFormatMarketSummary_Good_Minimal_Good(t *testing.T) {
	data := &coinData{
		Name:   "Unknown",
		Symbol: "ukn",
	}

	summary := FormatMarketSummary(data)
	assert.Contains(t, summary, "# Unknown (UKN)")
	// No price data, so these should be absent.
	assert.NotContains(t, summary, "Market Cap Rank")
}
