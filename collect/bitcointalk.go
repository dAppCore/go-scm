// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// FetchPageFunc is an injectable function type for fetching pages.
type FetchPageFunc func(ctx context.Context, url string) ([]btPost, error)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// SetHTTPClient replaces the package-level HTTP client.
func SetHTTPClient(c *http.Client) {
	if c != nil {
		httpClient = c
	}
}

// BitcoinTalkCollector collects forum posts from BitcoinTalk.
type BitcoinTalkCollector struct {
	TopicID string
	URL     string
	Pages   int
}

// BitcoinTalkCollectorWithFetcher wraps BitcoinTalkCollector with a custom fetcher for testing.
type BitcoinTalkCollectorWithFetcher struct {
	BitcoinTalkCollector
	Fetcher FetchPageFunc
}

type btPost struct {
	Number  int
	Author  string
	Date    string
	Content string
}

func (b *BitcoinTalkCollector) Name() string { return "bitcointalk" }

func (b *BitcoinTalkCollectorWithFetcher) Name() string { return b.BitcoinTalkCollector.Name() }

// Collect gathers posts from a BitcoinTalk topic.
func (b *BitcoinTalkCollector) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	if cfg == nil {
		return nil, errors.New("collect.BitcoinTalkCollector.Collect: config is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	topicID := b.TopicID
	if topicID == "" && b.URL != "" {
		topicID = extractBitcoinTalkTopicID(b.URL)
	}
	if topicID == "" {
		return &Result{Source: b.Name()}, nil
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(b.Name(), "Starting BitcoinTalk collection")
	}

	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(b.Name(), fmt.Sprintf("[dry-run] Would collect topic %s", topicID), nil)
			cfg.Dispatcher.EmitComplete(b.Name(), "BitcoinTalk dry-run complete", &Result{Source: b.Name()})
		}
		return &Result{Source: b.Name()}, nil
	}

	result := &Result{Source: b.Name()}
	pages := b.Pages
	if pages < 0 {
		pages = 0
	}
	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if pages > 0 && page > pages {
			break
		}
		if cfg.Limiter != nil {
			if err := cfg.Limiter.Wait(ctx, b.Name()); err != nil {
				return result, err
			}
		}
		url := b.pageURL(topicID, page)
		html, err := b.fetchPage(ctx, url)
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(b.Name(), fmt.Sprintf("Failed to fetch page %d: %v", page, err), nil)
			}
			break
		}
		posts, err := ParsePostsFromHTML(html)
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(b.Name(), fmt.Sprintf("Failed to parse page %d: %v", page, err), nil)
			}
			break
		}
		if len(posts) == 0 {
			break
		}
		for _, post := range posts {
			result.Items++
			md := FormatPostMarkdown(post.Number, post.Author, post.Date, post.Content)
			name := fmt.Sprintf("%s-page-%d-post-%d.md", topicID, page, post.Number)
			outPath, err := writeResultFile(cfg, b.Name(), name, md)
			if err != nil {
				result.Errors++
				continue
			}
			result.Files = append(result.Files, outPath)
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitItem(b.Name(), fmt.Sprintf("Post %d by %s", post.Number, post.Author), nil)
			}
		}
		if pages == 0 {
			// Continue until the source runs out of pages.
			page++
			continue
		}
		page++
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(b.Name(), fmt.Sprintf("Collected %d posts", result.Items), result)
	}
	return result, nil
}

// Collect gathers posts from a BitcoinTalk topic using the injected fetcher.
func (b *BitcoinTalkCollectorWithFetcher) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	if b.Fetcher == nil {
		return b.BitcoinTalkCollector.Collect(ctx, cfg)
	}
	if cfg == nil {
		return nil, errors.New("collect.BitcoinTalkCollectorWithFetcher.Collect: config is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	topicID := b.TopicID
	if topicID == "" && b.URL != "" {
		topicID = extractBitcoinTalkTopicID(b.URL)
	}
	if topicID == "" {
		return &Result{Source: b.Name()}, nil
	}
	if cfg.DryRun {
		return &Result{Source: b.Name()}, nil
	}
	result := &Result{Source: b.Name()}
	pages := b.Pages
	if pages < 0 {
		pages = 0
	}
	page := 1
	for {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if pages > 0 && page > pages {
			break
		}
		if cfg.Limiter != nil {
			if err := cfg.Limiter.Wait(ctx, b.Name()); err != nil {
				return result, err
			}
		}
		url := b.pageURL(topicID, page)
		posts, err := b.Fetcher(ctx, url)
		if err != nil {
			result.Errors++
			break
		}
		if len(posts) == 0 {
			break
		}
		for _, post := range posts {
			result.Items++
			md := FormatPostMarkdown(post.Number, post.Author, post.Date, post.Content)
			outPath, err := writeResultFile(cfg, b.Name(), fmt.Sprintf("%s-%d.md", topicID, post.Number), md)
			if err != nil {
				result.Errors++
				continue
			}
			result.Files = append(result.Files, outPath)
		}
		if pages == 0 {
			page++
			continue
		}
		page++
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(b.Name(), fmt.Sprintf("Collected %d posts", result.Items), result)
	}
	return result, nil
}

func (b *BitcoinTalkCollector) pageURL(topicID string, page int) string {
	base := b.URL
	if base == "" {
		base = "https://bitcointalk.org/index.php?topic=" + topicID + ".0"
	}
	if page <= 1 {
		return base
	}
	if strings.Contains(base, ".0") {
		return strings.Replace(base, ".0", "."+strconv.Itoa((page-1)*20), 1)
	}
	return base + "&page=" + strconv.Itoa(page)
}

func (b *BitcoinTalkCollector) fetchPage(ctx context.Context, url string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return fetchBitcoinTalkPage(ctx, url)
}

func fetchBitcoinTalkPage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("collect.BitcoinTalkCollector: http %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func extractBitcoinTalkTopicID(url string) string {
	re := regexp.MustCompile(`topic=(\d+)`)
	match := re.FindStringSubmatch(url)
	if len(match) == 2 {
		return match[1]
	}
	return ""
}

// ParsePostsFromHTML parses BitcoinTalk posts from raw HTML content.
func ParsePostsFromHTML(htmlContent string) ([]btPost, error) {
	if strings.TrimSpace(htmlContent) == "" {
		return nil, nil
	}
	root, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return parsePostsFallback(htmlContent), nil
	}

	var posts []btPost
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "post") {
			post := btPost{Number: len(posts) + 1}
			post.Author = findTextByClass(n, "author")
			post.Date = findTextByClass(n, "date")
			post.Content = strings.TrimSpace(renderTextFragment(n))
			if post.Content == "" {
				post.Content = strings.TrimSpace(textContent(n))
			}
			posts = append(posts, post)
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	if len(posts) > 0 {
		return posts, nil
	}

	plain := strings.TrimSpace(stripTags(htmlContent))
	if plain == "" {
		return nil, nil
	}
	return []btPost{{Number: 1, Content: plain}}, nil
}

// FormatPostMarkdown formats a post as markdown.
func FormatPostMarkdown(num int, author, date, content string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Post %d\n\n", num)
	if author != "" {
		fmt.Fprintf(&b, "- Author: %s\n", author)
	}
	if date != "" {
		fmt.Fprintf(&b, "- Date: %s\n", date)
	}
	b.WriteString("\n")
	b.WriteString(strings.TrimSpace(content))
	b.WriteString("\n")
	return b.String()
}

func stripTags(input string) string {
	re := regexp.MustCompile(`(?is)<[^>]+>`)
	return strings.TrimSpace(re.ReplaceAllString(input, " "))
}

func extractTagText(block, name string) string {
	re := regexp.MustCompile(`(?is)<[^>]*class="[^"]*` + regexp.QuoteMeta(name) + `[^"]*"[^>]*>(.*?)</[^>]+>`)
	match := re.FindStringSubmatch(block)
	if len(match) == 2 {
		return strings.TrimSpace(stripTags(match[1]))
	}
	return ""
}

func parsePostsFallback(htmlContent string) []btPost {
	re := regexp.MustCompile(`(?is)<div[^>]*class="[^"]*post[^"]*"[^>]*>(.*?)</div>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)
	if len(matches) == 0 {
		plain := strings.TrimSpace(stripTags(htmlContent))
		if plain == "" {
			return nil
		}
		return []btPost{{Number: 1, Content: plain}}
	}
	posts := make([]btPost, 0, len(matches))
	for i, match := range matches {
		block := match[1]
		posts = append(posts, btPost{
			Number:  i + 1,
			Author:  extractTagText(block, "author"),
			Date:    extractTagText(block, "date"),
			Content: strings.TrimSpace(stripTags(block)),
		})
	}
	return posts
}

func hasClass(node *html.Node, class string) bool {
	for _, attr := range node.Attr {
		if attr.Key != "class" {
			continue
		}
		for _, part := range strings.Fields(attr.Val) {
			if strings.EqualFold(part, class) {
				return true
			}
		}
	}
	return false
}

func findTextByClass(node *html.Node, class string) string {
	if node == nil {
		return ""
	}
	var found string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil || found != "" {
			return
		}
		if n.Type == html.ElementNode && hasClass(n, class) {
			found = strings.TrimSpace(renderTextFragment(n))
			if found != "" {
				return
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return found
}

func renderTextFragment(node *html.Node) string {
	if node == nil {
		return ""
	}
	var buf bytes.Buffer
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := html.Render(&buf, child); err != nil {
			return textContent(node)
		}
	}
	return stripTags(buf.String())
}

func textContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.TrimSpace(b.String())
}
