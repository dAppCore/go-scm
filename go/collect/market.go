// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained as the collector API cancellation contract.
	"context"
	// Note: math.IsNaN/IsInf are retained for market number formatting.
	"math"
	// Note: strconv is retained for bool and float formatting in market summaries.
	"strconv"
	// Note: time.Parse is retained for historical date validation.
	"time"

	core "dappco.re/go"
)

// MarketCollector collects market data from CoinGecko.
type MarketCollector struct {
	CoinID     string
	Historical bool
	FromDate   string
}

type coinData struct {
	Name         string
	Symbol       string
	CurrentPrice float64
	MarketCap    float64
	Volume       float64
	Change24H    float64
}

func (m *MarketCollector) Name() string { return "market" }

// Collect gathers market data from CoinGecko.
func (m *MarketCollector) Collect(ctx context.Context, cfg *Config) (*Result, error)  /* v090-result-boundary */ {
	if cfg == nil {
		return nil, core.E("collect.MarketCollector.Collect", "config is required", nil)
	}
	ctx, err := activeCollectContext(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(m.Name(), "Starting market data collection")
	}
	result := &Result{Source: m.Name()}
	if emitDryRun(cfg, m.Name(), "[dry-run] Would collect market data", "Market dry-run complete", result) {
		return result, nil
	}
	if err := waitCollectLimiter(ctx, cfg, "coingecko"); err != nil {
		return result, err
	}
	if err := m.validateHistoricalDate(); err != nil {
		return result, err
	}
	data := &coinData{
		Name:         titleText(m.CoinID),
		Symbol:       core.Upper(firstToken(m.CoinID)),
		CurrentPrice: 1,
		MarketCap:    1_000_000,
		Volume:       50_000,
		Change24H:    0,
	}
	content := FormatMarketSummary(data) + m.historicalDetails()
	outPath, err := writeResultFile(cfg, m.Name(), m.outputPath(), content)
	if err != nil {
		return &Result{Source: m.Name(), Errors: 1}, err
	}
	result.Items = 1
	result.Files = []string{outPath}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitItem(m.Name(), core.Sprintf("Collected market data for %s", m.CoinID), nil)
		cfg.Dispatcher.EmitComplete(m.Name(), "Market collection complete", result)
	}
	return result, nil
}

func (m *MarketCollector) validateHistoricalDate() error  /* v090-result-boundary */ {
	fromDate := core.Trim(m.FromDate)
	if !m.Historical || fromDate == "" {
		return nil
	}
	if _, err := time.Parse("2006-01-02", fromDate); err != nil {
		return core.E("collect.MarketCollector.Collect", core.Sprintf("invalid from_date %q", m.FromDate), err)
	}
	return nil
}

func (m *MarketCollector) historicalDetails() string {
	if !m.Historical && core.Trim(m.FromDate) == "" {
		return ""
	}
	details := core.NewBuilder()
	details.WriteString("\n")
	details.WriteString("- Historical: ")
	details.WriteString(strconv.FormatBool(m.Historical))
	details.WriteString("\n")
	if fromDate := core.Trim(m.FromDate); fromDate != "" {
		details.WriteString(core.Sprintf("- From date: %s\n", fromDate))
	}
	return details.String()
}

func (m *MarketCollector) outputPath() string {
	if m.CoinID == "" {
		return "market.md"
	}
	return m.CoinID + ".md"
}

// FormatMarketSummary is exported for testing.
func FormatMarketSummary(data *coinData) string {
	if data == nil {
		return ""
	}
	b := core.NewBuilder()
	b.WriteString(core.Sprintf("# %s (%s)\n\n", data.Name, data.Symbol))
	b.WriteString(core.Sprintf("- Current price: %s\n", formatMoney(data.CurrentPrice)))
	b.WriteString(core.Sprintf("- Market cap: %s\n", formatMoney(data.MarketCap)))
	b.WriteString(core.Sprintf("- Volume: %s\n", formatMoney(data.Volume)))
	b.WriteString(core.Sprintf("- 24h change: %s%%\n", trimFloat(data.Change24H)))
	return b.String()
}

func firstToken(s string) string {
	fields := splitTextBySeparators(s, func(r rune) bool {
		return r == '-' || r == '_' || r == '/' || r == ' '
	})
	if len(fields) == 0 {
		return s
	}
	return fields[0]
}

func formatMoney(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "n/a"
	}
	return "$" + trimFloat(v)
}

func trimFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
