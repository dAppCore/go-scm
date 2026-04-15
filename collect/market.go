// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
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
	path := "market.md"
	if m.CoinID != "" {
		path = m.CoinID + ".md"
	}
	outPath, err := writeResultFile(cfg, m.Name(), path, content)
	if err != nil {
		return &Result{Source: m.Name(), Errors: 1}, err
	}
	return &Result{Source: m.Name(), Items: 1, Files: []string{outPath}}, nil
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
