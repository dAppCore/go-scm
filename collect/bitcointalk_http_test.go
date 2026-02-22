package collect

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forge.lthn.ai/core/go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleBTCTalkPage returns HTML resembling a BitcoinTalk topic page with the
// given number of posts. If fewer than postsPerPage the caller can infer that
// it is the last page.
func sampleBTCTalkPage(count int) string {
	var page strings.Builder
	page.WriteString(`<html><body>`)
	for i := range count {
		page.WriteString(fmt.Sprintf(`
		<div class="post">
			<div class="poster_info">user%d</div>
			<div class="headerandpost">
				<div class="smalltext">January %02d, 2009</div>
			</div>
			<div class="inner">Post content number %d.</div>
		</div>`, i, i+1, i))
	}
	page.WriteString(`</body></html>`)
	return page.String()
}

func TestBitcoinTalkCollector_Collect_Good_OnePage(t *testing.T) {
	// Serve a single page with 5 posts (< 20, so collection stops after one page).
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleBTCTalkPage(5)))
	}))
	defer srv.Close()

	// Override the package-level HTTP client so requests go to our test server.
	oldClient := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = oldClient }()

	// We also need to redirect the URL that fetchPage constructs.
	// The easiest approach: use SetHTTPClient with a custom transport.
	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	httpClient = &http.Client{Transport: transport}

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil // disable rate limiting for tests

	b := &BitcoinTalkCollector{TopicID: "12345"}
	result, err := b.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 5, result.Items)
	assert.Equal(t, 0, result.Errors)
	assert.Len(t, result.Files, 5)
	assert.Equal(t, "bitcointalk:12345", result.Source)

	// Verify files were written.
	for i := 1; i <= 5; i++ {
		path := fmt.Sprintf("/output/bitcointalk/12345/posts/%d.md", i)
		content, err := m.Read(path)
		require.NoError(t, err, "file %s should exist", path)
		assert.Contains(t, content, fmt.Sprintf("Post %d by", i))
	}
}

func TestBitcoinTalkCollector_Collect_Good_PageLimit(t *testing.T) {
	pageCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pageCount++
		w.Header().Set("Content-Type", "text/html")
		// Return a full page (20 posts) each time so collection would continue
		// indefinitely without a Pages limit.
		_, _ = w.Write([]byte(sampleBTCTalkPage(20)))
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	b := &BitcoinTalkCollector{TopicID: "99999", Pages: 2}
	result, err := b.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 40, result.Items) // 2 pages * 20 posts
	assert.Equal(t, 2, pageCount)
}

func TestBitcoinTalkCollector_Collect_Good_CancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleBTCTalkPage(5)))
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	b := &BitcoinTalkCollector{TopicID: "12345"}
	_, err := b.Collect(ctx, cfg)
	assert.Error(t, err)
}

func TestBitcoinTalkCollector_Collect_Bad_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	b := &BitcoinTalkCollector{TopicID: "12345"}
	result, err := b.Collect(context.Background(), cfg)

	// fetchPage error causes break with Errors incremented.
	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
	assert.Equal(t, 1, result.Errors)
}

func TestBitcoinTalkCollector_Collect_Good_EmitsEvents(t *testing.T) {
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

	var starts, items, completes int
	cfg.Dispatcher.On(EventStart, func(e Event) { starts++ })
	cfg.Dispatcher.On(EventItem, func(e Event) { items++ })
	cfg.Dispatcher.On(EventComplete, func(e Event) { completes++ })

	b := &BitcoinTalkCollector{TopicID: "12345"}
	result, err := b.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
	assert.Equal(t, 1, starts)
	assert.Equal(t, 2, items)
	assert.Equal(t, 1, completes)
}

func TestSetHTTPClient_Good(t *testing.T) {
	old := httpClient
	defer func() { httpClient = old }()

	custom := &http.Client{}
	SetHTTPClient(custom)
	assert.Equal(t, custom, httpClient)
}

func TestFetchPage_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(sampleBTCTalkPage(3)))
	}))
	defer srv.Close()

	old := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = old }()

	b := &BitcoinTalkCollector{TopicID: "12345"}
	posts, err := b.fetchPage(context.Background(), srv.URL)

	require.NoError(t, err)
	assert.Len(t, posts, 3)
}

func TestFetchPage_Bad_StatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	old := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = old }()

	b := &BitcoinTalkCollector{TopicID: "12345"}
	_, err := b.fetchPage(context.Background(), srv.URL)
	assert.Error(t, err)
}

func TestFetchPage_Bad_InvalidHTML(t *testing.T) {
	// html.Parse is very forgiving, so serve an empty page.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html><body></body></html>`))
	}))
	defer srv.Close()

	old := httpClient
	httpClient = srv.Client()
	defer func() { httpClient = old }()

	b := &BitcoinTalkCollector{TopicID: "12345"}
	posts, err := b.fetchPage(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Empty(t, posts)
}

// rewriteTransport rewrites all request URLs to point at the test server.
type rewriteTransport struct {
	base   http.RoundTripper
	target string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = "http"
	req.URL.Host = t.target[len("http://"):]
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
