package collect

import (
	"context"
	"testing"

	"forge.lthn.ai/core/go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLToMarkdown_Good_OrderedList(t *testing.T) {
	input := `<ol><li>First</li><li>Second</li><li>Third</li></ol>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "1. First")
	assert.Contains(t, result, "2. Second")
	assert.Contains(t, result, "3. Third")
}

func TestHTMLToMarkdown_Good_UnorderedList(t *testing.T) {
	input := `<ul><li>Alpha</li><li>Beta</li></ul>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "- Alpha")
	assert.Contains(t, result, "- Beta")
}

func TestHTMLToMarkdown_Good_Blockquote(t *testing.T) {
	input := `<blockquote>A wise quote</blockquote>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "> A wise quote")
}

func TestHTMLToMarkdown_Good_HorizontalRule(t *testing.T) {
	input := `<p>Before</p><hr/><p>After</p>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "---")
}

func TestHTMLToMarkdown_Good_LinkWithoutHref(t *testing.T) {
	input := `<a>bare link text</a>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "bare link text")
	assert.NotContains(t, result, "[")
}

func TestHTMLToMarkdown_Good_H4H5H6(t *testing.T) {
	input := `<h4>H4</h4><h5>H5</h5><h6>H6</h6>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "#### H4")
	assert.Contains(t, result, "##### H5")
	assert.Contains(t, result, "###### H6")
}

func TestHTMLToMarkdown_Good_StripsStyle(t *testing.T) {
	input := `<html><head><style>.foo{color:red}</style></head><body><p>Clean</p></body></html>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "Clean")
	assert.NotContains(t, result, "color")
}

func TestHTMLToMarkdown_Good_LineBreak(t *testing.T) {
	input := `<p>Line one<br/>Line two</p>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "Line one")
	assert.Contains(t, result, "Line two")
}

func TestHTMLToMarkdown_Good_NestedBoldItalic(t *testing.T) {
	input := `<b>bold text</b> and <i>italic text</i>`
	result, err := HTMLToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "**bold text**")
	assert.Contains(t, result, "*italic text*")
}

func TestJSONToMarkdown_Good_NestedObject(t *testing.T) {
	input := `{"outer": {"inner_key": "inner_value"}}`
	result, err := JSONToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "**outer:**")
	assert.Contains(t, result, "**inner_key:** inner_value")
}

func TestJSONToMarkdown_Good_NestedArray(t *testing.T) {
	input := `[["a", "b"], ["c"]]`
	result, err := JSONToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "# Data")
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
}

func TestJSONToMarkdown_Good_ScalarValue(t *testing.T) {
	input := `42`
	result, err := JSONToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "42")
}

func TestJSONToMarkdown_Good_ArrayOfObjects(t *testing.T) {
	input := `[{"name": "Alice"}, {"name": "Bob"}]`
	result, err := JSONToMarkdown(input)
	require.NoError(t, err)
	assert.Contains(t, result, "Item 1")
	assert.Contains(t, result, "Alice")
	assert.Contains(t, result, "Item 2")
	assert.Contains(t, result, "Bob")
}

func TestProcessor_Process_Good_CancelledContext(t *testing.T) {
	m := io.NewMockMedium()
	m.Dirs["/input"] = true
	m.Files["/input/file.html"] = `<h1>Test</h1>`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := &Processor{Source: "test", Dir: "/input"}
	_, err := p.Process(ctx, cfg)
	assert.Error(t, err)
}

func TestProcessor_Process_Good_EmitsEvents(t *testing.T) {
	m := io.NewMockMedium()
	m.Dirs["/input"] = true
	m.Files["/input/a.html"] = `<h1>Title</h1>`
	m.Files["/input/b.json"] = `{"key": "value"}`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	var starts, items, completes int
	cfg.Dispatcher.On(EventStart, func(e Event) { starts++ })
	cfg.Dispatcher.On(EventItem, func(e Event) { items++ })
	cfg.Dispatcher.On(EventComplete, func(e Event) { completes++ })

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Items)
	assert.Equal(t, 1, starts)
	assert.Equal(t, 2, items)
	assert.Equal(t, 1, completes)
}

func TestProcessor_Process_Good_BadHTML(t *testing.T) {
	m := io.NewMockMedium()
	m.Dirs["/input"] = true
	// html.Parse is very tolerant, so even bad HTML will parse. But we test
	// that the pipeline handles it gracefully.
	m.Files["/input/bad.html"] = `<html><body><p>Still valid enough</p>`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Items)
}

func TestProcessor_Process_Good_BadJSON(t *testing.T) {
	m := io.NewMockMedium()
	m.Dirs["/input"] = true
	m.Files["/input/bad.json"] = `not valid json`

	cfg := NewConfigWithMedium(m, "/output")
	cfg.Limiter = nil

	var errors int
	cfg.Dispatcher.On(EventError, func(e Event) { errors++ })

	p := &Processor{Source: "test", Dir: "/input"}
	result, err := p.Process(context.Background(), cfg)

	require.NoError(t, err)
	assert.Equal(t, 0, result.Items)
	assert.Equal(t, 1, result.Errors)
	assert.Equal(t, 1, errors)
}
