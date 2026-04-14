// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dappco.re/go/core/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- GitHub collector: context cancellation and orchestration ---

func TestGitHubCollector_Collect_Good_ContextCancelledInLoop_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = false

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	g := &GitHubCollector{Org: "test-org", Repo: "test-repo"}
	result, err := g.Collect(ctx, cfg)

	// The context cancellation should be detected in the loop
	assert.Error(t, err)
	assert.NotNil(t, result)
}

func TestGitHubCollector_Collect_Good_IssuesOnlyDryRunProgress_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	var progressCount int
	cfg.Dispatcher.On(EventProgress, func(e Event) { progressCount++ })

	g := &GitHubCollector{Org: "test-org", Repo: "test-repo", IssuesOnly: true}
	result, err := g.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
	assert.GreaterOrEqual(t, progressCount, 1)
}

func TestGitHubCollector_Collect_Good_PRsOnlyDryRunSkipsIssues_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	g := &GitHubCollector{Org: "test-org", Repo: "test-repo", PRsOnly: true}
	result, err := g.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestGitHubCollector_Collect_Good_EmitsStartAndComplete_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	var starts, completes int
	cfg.Dispatcher.On(EventStart, func(e Event) { starts++ })
	cfg.Dispatcher.On(EventComplete, func(e Event) { completes++ })

	g := &GitHubCollector{Org: "test-org", Repo: "test-repo"}
	_, err := g.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 1, starts)
	assert.Equal(t, 1, completes)
}

func TestGitHubCollector_Collect_Good_NilDispatcherHandled_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true
	cfg.Dispatcher = nil

	g := &GitHubCollector{Org: "test-org", Repo: "test-repo"}
	result, err := g.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestFormatIssueMarkdown_Good_NoBodyNoURL_Good(t *testing.T) {
	issue := ghIssue{
		Number: 1,
		Title:  "No Body Issue",
		State:  "open",
		Author: ghAuthor{Login: "user"},
		URL:    "",
		Body:   "",
	}

	md := formatIssueMarkdown(issue)
	assert.Contains(t, md, "# No Body Issue")
	assert.NotContains(t, md, "**URL:**")
}

// --- Market collector: fetchJSON edge cases ---

func TestFetchJSON_Bad_NonJSONBody_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html>not json</html>`))
	}))
	defer srv.Close()

	_, err := fetchJSON[coinData](context.Background(), srv.URL)
	assert.Error(t, err)
}

func TestFetchJSON_Bad_MalformedURL_Good(t *testing.T) {
	_, err := fetchJSON[coinData](context.Background(), "://bad-url")
	assert.Error(t, err)
}

func TestFetchJSON_Bad_ServerUnavailable_Good(t *testing.T) {
	_, err := fetchJSON[coinData](context.Background(), "http://127.0.0.1:1")
	assert.Error(t, err)
}

func TestFetchJSON_Bad_Non200StatusCode_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := fetchJSON[coinData](context.Background(), srv.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code")
}

func TestMarketCollector_Collect_Bad_MissingCoinID_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	mc := &MarketCollector{CoinID: ""}
	_, err := mc.Collect(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "coin ID is required")
}

func TestMarketCollector_Collect_Good_NoDispatcher_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := coinData{ID: "test", Symbol: "tst", Name: "Test",
			MarketData: marketData{CurrentPrice: map[string]float64{"usd": 1.0}}}
		_ = json.NewEncoder(w).Encode(data)
	}))
	defer srv.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = srv.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil
	cfg.Dispatcher = nil

	mc := &MarketCollector{CoinID: "test"}
	result, err := mc.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
}

func TestMarketCollector_Collect_Bad_CurrentFetchFails_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = srv.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	mc := &MarketCollector{CoinID: "fail-coin"}
	result, err := mc.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
	assert.Equal(t, 1, result.Errors)
}

func TestMarketCollector_CollectHistorical_Good_DefaultDays_Good(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			data := coinData{ID: "test", Symbol: "tst", Name: "Test",
				MarketData: marketData{CurrentPrice: map[string]float64{"usd": 1.0}}}
			_ = json.NewEncoder(w).Encode(data)
		} else {
			assert.Contains(t, r.URL.RawQuery, "days=365")
			data := historicalData{Prices: [][]float64{{1705305600000, 1.0}}}
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

	mc := &MarketCollector{CoinID: "test", Historical: true}
	result, err := mc.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 3, result.Items)
}

func TestMarketCollector_CollectHistorical_Good_WithRateLimiter_Good(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			data := coinData{ID: "test", Symbol: "tst", Name: "Test",
				MarketData: marketData{CurrentPrice: map[string]float64{"usd": 1.0}}}
			_ = json.NewEncoder(w).Encode(data)
		} else {
			data := historicalData{Prices: [][]float64{{1705305600000, 1.0}}}
			_ = json.NewEncoder(w).Encode(data)
		}
	}))
	defer srv.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = srv.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = NewRateLimiter()
	cfg.Limiter.SetDelay("coingecko", 1*time.Millisecond)

	mc := &MarketCollector{CoinID: "test", Historical: true}
	result, err := mc.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 3, result.Items)
}

// --- State: error paths ---

func TestState_Load_Bad_MalformedJSON_Good(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/state.json"] = `{invalid json`

	s := NewState(m, "/state.json")
	err := s.Load()
	assert.Error(t, err)
}

// --- Process: additional coverage for uncovered branches ---

func TestHTMLToMarkdown_Good_PreCodeBlock_Good(t *testing.T) {
	input := `<pre>some code here</pre>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "```")
	assert.Contains(t, result, "some code here")
}

func TestHTMLToMarkdown_Good_StrongAndEmElements_Good(t *testing.T) {
	input := `<strong>bold</strong> and <em>italic</em>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "**bold**")
	assert.Contains(t, result, "*italic*")
}

func TestHTMLToMarkdown_Good_InlineCode_Good(t *testing.T) {
	input := `<code>var x = 1</code>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "`var x = 1`")
}

func TestHTMLToMarkdown_Good_AnchorWithHref_Good(t *testing.T) {
	input := `<a href="https://example.com">Click here</a>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "[Click here](https://example.com)")
}

func TestHTMLToMarkdown_Good_ScriptTagRemoved_Good(t *testing.T) {
	input := `<html><body><script>alert('xss')</script><p>Safe text</p></body></html>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "Safe text")
	assert.NotContains(t, result, "alert")
}

func TestHTMLToMarkdown_Good_H1H2H3Headers_Good(t *testing.T) {
	input := `<h1>One</h1><h2>Two</h2><h3>Three</h3>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "# One")
	assert.Contains(t, result, "## Two")
	assert.Contains(t, result, "### Three")
}

func TestHTMLToMarkdown_Good_MultiParagraph_Good(t *testing.T) {
	input := `<p>First paragraph</p><p>Second paragraph</p>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "First paragraph")
	assert.Contains(t, result, "Second paragraph")
}

func TestJSONToMarkdown_Bad_Malformed_Good(t *testing.T) {
	_, err := JSONToMarkdown(`{invalid}`)
	assert.Error(t, err)
}

func TestJSONToMarkdown_Good_FlatObject_Good(t *testing.T) {
	input := `{"name": "Alice", "age": 30}`
	result, err := JSONToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "**name:** Alice")
	assert.Contains(t, result, "**age:** 30")
}

func TestJSONToMarkdown_Good_ScalarList_Good(t *testing.T) {
	input := `["hello", "world"]`
	result, err := JSONToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "- hello")
	assert.Contains(t, result, "- world")
}

func TestJSONToMarkdown_Good_ObjectContainingArray_Good(t *testing.T) {
	input := `{"items": [1, 2, 3]}`
	result, err := JSONToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "**items:**")
}

func TestProcessor_Process_Bad_MissingDir_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")

	p := &Processor{Source: "test", Dir: ""}
	_, err := p.Process(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory is required")
}

func TestProcessor_Process_Good_DryRunEmitsProgress_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	var progressCount int
	cfg.Dispatcher.On(EventProgress, func(e Event) { progressCount++ })

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
	assert.Equal(t, 1, progressCount)
}

func TestProcessor_Process_Good_SkipsUnsupportedExtension_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/input")
	m.Files["/input/data.csv"] = `a,b,c`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
	assert.Equal(t, 1, result.Skipped)
}

func TestProcessor_Process_Good_MarkdownPassthroughTrimmed_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/input")
	m.Files["/input/readme.md"] = `# Hello World  `

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Items)

	content, readErr := m.Read("/output/processed/test/readme.md")
	require.NoError(t, readErr)
	assert.Equal(t, "# Hello World", content)
}

func TestProcessor_Process_Good_HTMExtensionHandled_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/input")
	m.Files["/input/page.htm"] = `<h1>HTM File</h1>`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Items)
}

func TestProcessor_Process_Good_NilDispatcherHandled_Good(t *testing.T) {
	m := io.NewMockMedium()
	_ = m.EnsureDir("/input")
	m.Files["/input/test.html"] = `<p>Text</p>`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil
	cfg.Dispatcher = nil

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Items)
}

// --- BitcoinTalk: additional edge cases ---

func TestBitcoinTalkCollector_Name_Good_EmptyTopicAndURL_Good(t *testing.T) {
	b := &BitcoinTalkCollector{}
	assert.Equal(t, "bitcointalk:", b.Name())
}

func TestBitcoinTalkCollector_Collect_Good_NilDispatcherHandled_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleBTCTalkPage(2)))
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil
	cfg.Dispatcher = nil

	b := &BitcoinTalkCollector{TopicID: "12345"}
	result, err := b.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
}

func TestBitcoinTalkCollector_Collect_Good_DryRunEmitsProgress_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	var progressEmitted bool
	cfg.Dispatcher.On(EventProgress, func(e Event) { progressEmitted = true })

	b := &BitcoinTalkCollector{TopicID: "12345"}
	result, err := b.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
	assert.True(t, progressEmitted)
}

func TestParsePostsFromHTML_Good_PostWithNoInnerContent_Good(t *testing.T) {
	htmlContent := `<html><body>
		<div class="post">
			<div class="poster_info">user1</div>
		</div>
	</body></html>`
	posts, err := ParsePostsFromHTML(htmlContent)
	require.NoError(t, err)
	assert.Empty(t, posts)
}

func TestFormatPostMarkdown_Good_WithDateContent_Good(t *testing.T) {
	md := FormatPostMarkdown(1, "alice", "2025-01-15", "Hello world")
	assert.Contains(t, md, "# Post 1 by alice")
	assert.Contains(t, md, "**Date:** 2025-01-15")
	assert.Contains(t, md, "Hello world")
}

// --- Papers collector: edge cases ---

func TestPapersCollector_Collect_Good_DryRunEmitsProgress_Good(t *testing.T) {
	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.DryRun = true

	var progressEmitted bool
	cfg.Dispatcher.On(EventProgress, func(e Event) { progressEmitted = true })

	p := &PapersCollector{Source: PaperSourceIACR, Query: "test"}
	result, err := p.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
	assert.True(t, progressEmitted)
}

func TestPapersCollector_Collect_Good_NilDispatcherIACR_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleIACRHTML))
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil
	cfg.Dispatcher = nil

	p := &PapersCollector{Source: PaperSourceIACR, Query: "zero knowledge"}
	result, err := p.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
}

func TestArXivEntryToPaper_Good_NoAlternateLink_Good(t *testing.T) {
	entry := arxivEntry{
		ID:    "http://arxiv.org/abs/2501.99999v1",
		Title: "No Alternate",
		Links: []arxivLink{
			{Href: "http://arxiv.org/pdf/2501.99999v1", Rel: "related"},
		},
	}

	p := arxivEntryToPaper(entry)
	assert.Equal(t, "http://arxiv.org/abs/2501.99999v1", p.URL)
}

// --- Excavator: additional edge cases ---

func TestExcavator_Run_Good_ResumeLoadError_Good(t *testing.T) {
	m := io.NewMockMedium()
	m.Files["/output/.collect-state.json"] = `{invalid`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	c1 := &mockCollector{name: "source-a", items: 5}
	e := &Excavator{
		Collectors: []Collector{c1},
		Resume:     true,
	}

	_, err := e.Run(context.Background(), cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load state")
}

// --- RateLimiter: additional edge cases ---

func TestRateLimiter_Wait_Good_QuickSuccessiveCallsAfterDelay_Good(t *testing.T) {
	rl := NewRateLimiter()
	rl.SetDelay("fast", 1*time.Millisecond)

	ctx := context.Background()

	err := rl.Wait(ctx, "fast")
	assert.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	start := time.Now()
	err = rl.Wait(ctx, "fast")
	assert.NoError(t, err)
	assert.Less(t, time.Since(start), 5*time.Millisecond)
}

// --- FormatMarketSummary: with empty market data values ---

func TestFormatMarketSummary_Good_ZeroRank_Good(t *testing.T) {
	data := &coinData{
		Name:   "Tiny Token",
		Symbol: "tiny",
		MarketData: marketData{
			CurrentPrice:  map[string]float64{"usd": 0.0001},
			MarketCapRank: 0, // should not appear
		},
	}
	summary := FormatMarketSummary(data)
	assert.Contains(t, summary, "# Tiny Token (TINY)")
	assert.NotContains(t, summary, "Market Cap Rank")
}

func TestFormatMarketSummary_Good_ZeroSupply_Good(t *testing.T) {
	data := &coinData{
		Name:   "Zero Supply",
		Symbol: "zs",
		MarketData: marketData{
			CirculatingSupply: 0,
			TotalSupply:       0,
		},
	}
	summary := FormatMarketSummary(data)
	assert.NotContains(t, summary, "Circulating Supply")
	assert.NotContains(t, summary, "Total Supply")
}

func TestFormatMarketSummary_Good_NoLastUpdated_Good(t *testing.T) {
	data := &coinData{
		Name:   "No Update",
		Symbol: "nu",
	}
	summary := FormatMarketSummary(data)
	assert.NotContains(t, summary, "Last updated")
}
