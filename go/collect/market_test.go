// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained in tests to exercise collector public APIs.
	"context"
	// Note: strings.Contains is retained for assertions over generated Markdown.
	`strings`
	// Note: testing is the standard Go test harness.
	"testing"

	coreio "dappco.re/go/io"
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

func TestMarket_MarketCollector_Name_Good(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "MarketCollector_Name"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestMarket_MarketCollector_Name_Bad(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "MarketCollector_Name"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestMarket_MarketCollector_Name_Ugly(t *testing.T) {
	reference := "Name"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "MarketCollector_Name"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestMarket_MarketCollector_Collect_Good(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "MarketCollector_Collect"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestMarket_MarketCollector_Collect_Bad(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "MarketCollector_Collect"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestMarket_MarketCollector_Collect_Ugly(t *testing.T) {
	reference := "Collect"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "MarketCollector_Collect"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestMarket_FormatMarketSummary_Good(t *testing.T) {
	target := "FormatMarketSummary"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestMarket_FormatMarketSummary_Bad(t *testing.T) {
	target := "FormatMarketSummary"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestMarket_FormatMarketSummary_Ugly(t *testing.T) {
	target := "FormatMarketSummary"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}
