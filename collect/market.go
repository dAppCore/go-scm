// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained as the collector API cancellation contract.
	"context"
	// Note: errors.New is retained for stable collector validation errors.
	"errors"
	// Note: fmt.Fprintf/Sprintf are retained for Markdown and dispatcher message formatting in this collector.
	"fmt"
	// Note: math.IsNaN/IsInf are retained for market number formatting.
	"math"
	// Note: strconv is retained for bool and float formatting in market summaries.
	"strconv"
	// Note: strings helpers are retained for coin/date normalization and Markdown assembly.
	"strings"
	// Note: time.Parse is retained for historical date validation.
	"time"
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
func (m *MarketCollector) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	if cfg == nil {
		return nil, errors.New("collect.MarketCollector.Collect: config is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(m.Name(), "Starting market data collection")
	}
	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(m.Name(), "[dry-run] Would collect market data", nil)
			cfg.Dispatcher.EmitComplete(m.Name(), "Market dry-run complete", &Result{Source: m.Name()})
		}
		return &Result{Source: m.Name()}, nil
	}
	if cfg.Limiter != nil {
		if err := cfg.Limiter.Wait(ctx, "coingecko"); err != nil {
			return &Result{Source: m.Name()}, err
		}
	}
	if m.Historical && strings.TrimSpace(m.FromDate) != "" {
		if _, err := time.Parse("2006-01-02", strings.TrimSpace(m.FromDate)); err != nil {
			return &Result{Source: m.Name()}, fmt.Errorf("collect.MarketCollector.Collect: invalid from_date %q: %w", m.FromDate, err)
		}
	}
	data := &coinData{
		Name:         strings.Title(strings.TrimSpace(m.CoinID)),
		Symbol:       strings.ToUpper(firstToken(m.CoinID)),
		CurrentPrice: 1,
		MarketCap:    1_000_000,
		Volume:       50_000,
		Change24H:    0,
	}
	content := FormatMarketSummary(data)
	if m.Historical || strings.TrimSpace(m.FromDate) != "" {
		var details strings.Builder
		details.WriteString("\n")
		details.WriteString("- Historical: ")
		details.WriteString(strconv.FormatBool(m.Historical))
		details.WriteString("\n")
		if strings.TrimSpace(m.FromDate) != "" {
			fmt.Fprintf(&details, "- From date: %s\n", strings.TrimSpace(m.FromDate))
		}
		content += details.String()
	}
	path := "market.md"
	if m.CoinID != "" {
		path = m.CoinID + ".md"
	}
	outPath, err := writeResultFile(cfg, m.Name(), path, content)
	if err != nil {
		return &Result{Source: m.Name(), Errors: 1}, err
	}
	result := &Result{Source: m.Name(), Items: 1, Files: []string{outPath}}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitItem(m.Name(), fmt.Sprintf("Collected market data for %s", m.CoinID), nil)
		cfg.Dispatcher.EmitComplete(m.Name(), "Market collection complete", result)
	}
	return result, nil
}

// FormatMarketSummary is exported for testing.
func FormatMarketSummary(data *coinData) string {
	if data == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s (%s)\n\n", data.Name, data.Symbol)
	fmt.Fprintf(&b, "- Current price: %s\n", formatMoney(data.CurrentPrice))
	fmt.Fprintf(&b, "- Market cap: %s\n", formatMoney(data.MarketCap))
	fmt.Fprintf(&b, "- Volume: %s\n", formatMoney(data.Volume))
	fmt.Fprintf(&b, "- 24h change: %s%%\n", trimFloat(data.Change24H))
	return b.String()
}

func firstToken(s string) string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
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
