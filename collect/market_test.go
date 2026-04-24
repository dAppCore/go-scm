// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained in tests to exercise collector public APIs.
	"context"
	// Note: strings.Contains is retained for assertions over generated Markdown.
	"strings"
	// Note: testing is the standard Go test harness.
	"testing"

	coreio "dappco.re/go/core/io"
)

func TestMarketCollectorIncludesHistoricalDetails(t *testing.T) {
	medium := coreio.NewMockMedium()
	cfg := NewConfigWithMedium(medium, "out")

	result, err := (&MarketCollector{
		CoinID:     "bitcoin",
		Historical: true,
		FromDate:   "2024-01-01",
	}).Collect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if result == nil || result.Items != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}

	raw, ok := medium.Files["out/market/bitcoin.md"]
	if !ok {
		t.Fatalf("expected market output to be written")
	}
	if !strings.Contains(raw, "- Historical: true") {
		t.Fatalf("historical flag not written: %q", raw)
	}
	if !strings.Contains(raw, "- From date: 2024-01-01") {
		t.Fatalf("from date not written: %q", raw)
	}
}

func TestMarketCollectorRejectsInvalidHistoricalDate(t *testing.T) {
	cfg := NewConfigWithMedium(coreio.NewMockMedium(), "out")

	_, err := (&MarketCollector{
		CoinID:     "bitcoin",
		Historical: true,
		FromDate:   "2024-99-99",
	}).Collect(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected invalid historical date to fail")
	}
}
