// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: strings.Contains is retained for assertions over parsed Markdown/text output.
	"strings"
	// Note: testing is the standard Go test harness.
	"testing"
)

func TestParsePostsFromHTMLParsesNestedPostBlocks(t *testing.T) {
	html := `
		<html>
			<body>
				<div class="post">
					<div class="author"> satoshi </div>
					<div class="date"> 2009-01-03 </div>
					<div class="body">
						<p>Hello <strong>Bitcoin</strong></p>
						<div>Nested <a href="https://example.com">link</a></div>
					</div>
				</div>
				<div class="post">
					<div class="author"> hal </div>
					<div class="date"> 2009-01-12 </div>
					<div class="body">Second post</div>
				</div>
			</body>
		</html>`

	posts, err := ParsePostsFromHTML(html)
	if err != nil {
		t.Fatalf("parse posts: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
	if posts[0].Number != 1 || posts[0].Author != "satoshi" || posts[0].Date != "2009-01-03" {
		t.Fatalf("unexpected first post: %#v", posts[0])
	}
	if !strings.Contains(posts[0].Content, "Hello") || !strings.Contains(posts[0].Content, "Bitcoin") || !strings.Contains(posts[0].Content, "link") {
		t.Fatalf("unexpected first post content: %q", posts[0].Content)
	}
	if posts[1].Number != 2 || posts[1].Author != "hal" || posts[1].Date != "2009-01-12" {
		t.Fatalf("unexpected second post: %#v", posts[1])
	}
}

func TestParsePostsFromHTMLFallsBackToPlainText(t *testing.T) {
	posts, err := ParsePostsFromHTML("<p>plain text only</p>")
	if err != nil {
		t.Fatalf("parse plain text: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 fallback post, got %d", len(posts))
	}
	if posts[0].Content != "plain text only" {
		t.Fatalf("unexpected fallback content: %q", posts[0].Content)
	}
}
