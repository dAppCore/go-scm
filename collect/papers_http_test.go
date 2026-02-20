package collect

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forge.lthn.ai/core/go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

const sampleIACRHTML = `<html><body>
<div class="paperentry">
	<a href="/eprint/2025/001">Zero-Knowledge Proofs</a>
	<span class="author">Alice</span>
	<span class="author">Bob</span>
	<span class="date">2025-01-15</span>
	<p class="abstract">We present a novel construction for zero-knowledge proofs.</p>
</div>
<div class="paperentry">
	<a href="/eprint/2025/002">Lattice Cryptography</a>
	<span class="author">Charlie</span>
	<span class="date">2025-01-20</span>
	<p class="abstract">A survey of lattice-based cryptography.</p>
</div>
</body></html>`

const sampleArXivXML = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <entry>
    <id>http://arxiv.org/abs/2501.12345v1</id>
    <title>Ring Signatures Revisited</title>
    <summary>We propose an efficient ring signature scheme.</summary>
    <published>2025-01-10T00:00:00Z</published>
    <author><name>Alice</name></author>
    <author><name>David</name></author>
    <link href="http://arxiv.org/abs/2501.12345v1" rel="alternate"/>
  </entry>
  <entry>
    <id>http://arxiv.org/abs/2501.67890v1</id>
    <title>Post-Quantum Signatures</title>
    <summary>A new approach to post-quantum digital signatures.</summary>
    <published>2025-01-12T00:00:00Z</published>
    <author><name>Eve</name></author>
    <link href="http://arxiv.org/abs/2501.67890v1" rel="alternate"/>
  </entry>
</feed>`

func TestPapersCollector_CollectIACR_Good(t *testing.T) {
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

	p := &PapersCollector{Source: PaperSourceIACR, Query: "zero knowledge"}
	result, err := p.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
	assert.Len(t, result.Files, 2)

	// Verify content was written.
	content, err := m.Read("/output/papers/iacr/2025-001.md")
	require.NoError(t, err)
	assert.Contains(t, content, "Zero-Knowledge Proofs")
}

func TestPapersCollector_CollectArXiv_Good(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(sampleArXivXML))
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &PapersCollector{Source: PaperSourceArXiv, Query: "ring signatures"}
	result, err := p.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
	assert.Len(t, result.Files, 2)

	// Verify one of the papers.
	content, err := m.Read("/output/papers/arxiv/2501.12345v1.md")
	require.NoError(t, err)
	assert.Contains(t, content, "Ring Signatures Revisited")
	assert.Contains(t, content, "Alice")
}

func TestPapersCollector_CollectArXiv_Good_WithCategory(t *testing.T) {
	var capturedQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(sampleArXivXML))
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &PapersCollector{Source: PaperSourceArXiv, Query: "crypto", Category: "cs.CR"}
	_, err := p.Collect(context.Background(), cfg)
	require.NoError(t, err)
	assert.Contains(t, capturedQuery, "cat")
}

func TestPapersCollector_CollectAll_Good(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call is IACR
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(sampleIACRHTML))
		} else {
			// Second call is arXiv
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(sampleArXivXML))
		}
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &PapersCollector{Source: PaperSourceAll, Query: "cryptography"}
	result, err := p.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 4, result.Items) // 2 IACR + 2 arXiv
}

func TestPapersCollector_CollectIACR_Bad_ServerError(t *testing.T) {
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

	p := &PapersCollector{Source: PaperSourceIACR, Query: "test"}
	_, err := p.Collect(context.Background(), cfg)
	assert.Error(t, err)
}

func TestPapersCollector_CollectArXiv_Bad_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &PapersCollector{Source: PaperSourceArXiv, Query: "test"}
	_, err := p.Collect(context.Background(), cfg)
	assert.Error(t, err)
}

func TestPapersCollector_CollectArXiv_Bad_InvalidXML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`not xml at all`))
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &PapersCollector{Source: PaperSourceArXiv, Query: "test"}
	_, err := p.Collect(context.Background(), cfg)
	assert.Error(t, err)
}

func TestPapersCollector_CollectAll_Bad_BothFail(t *testing.T) {
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

	p := &PapersCollector{Source: PaperSourceAll, Query: "test"}
	_, err := p.Collect(context.Background(), cfg)
	assert.Error(t, err)
}

func TestPapersCollector_CollectAll_Good_OneFails(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// IACR fails
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// ArXiv succeeds
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(sampleArXivXML))
		}
	}))
	defer srv.Close()

	transport := &rewriteTransport{base: srv.Client().Transport, target: srv.URL}
	old := httpClient
	httpClient = &http.Client{Transport: transport}
	defer func() { httpClient = old }()

	m := io.NewMockMedium()
	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &PapersCollector{Source: PaperSourceAll, Query: "test"}
	result, err := p.Collect(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
	assert.Equal(t, 1, result.Errors) // IACR failure counted
}

func TestExtractIACRPapers_Good(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(sampleIACRHTML))
	require.NoError(t, err)

	papers := extractIACRPapers(doc)
	assert.Len(t, papers, 2)

	assert.Equal(t, "Zero-Knowledge Proofs", papers[0].Title)
	assert.Contains(t, papers[0].Authors, "Alice")
	assert.Contains(t, papers[0].Authors, "Bob")
	assert.Equal(t, "2025-01-15", papers[0].Date)
	assert.Contains(t, papers[0].Abstract, "zero-knowledge proofs")
	assert.Equal(t, "iacr", papers[0].Source)

	assert.Equal(t, "Lattice Cryptography", papers[1].Title)
}

func TestExtractIACRPapers_Good_Empty(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`<html><body></body></html>`))
	require.NoError(t, err)

	papers := extractIACRPapers(doc)
	assert.Empty(t, papers)
}

func TestExtractIACRPapers_Good_NoTitle(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`<html><body><div class="paperentry"></div></body></html>`))
	require.NoError(t, err)

	papers := extractIACRPapers(doc)
	// Entry with no title should be excluded by the Title check.
	assert.Empty(t, papers)
}
