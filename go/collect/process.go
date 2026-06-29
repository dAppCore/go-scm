// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained as the processor API cancellation contract.
	"context"
	// Note: regexp is retained for HTML conversion patterns; no core equivalent covers compiled regexes.
	"regexp"

	core "dappco.re/go"
)

var (
	htmlAnchorRe = regexp.MustCompile(`(?is)<a[^>]*href=(?:"([^"]+)"|'([^']+)')[^>]*>(.*?)</a>`)
	htmlTagRe    = regexp.MustCompile(`(?is)<[^>]+>`)
)

// Processor converts collected data to clean markdown.
type Processor struct {
	Source string
	Dir    string
}

func (p *Processor) Name() string { return "process" }

// Process reads files from the source directory, converts HTML or JSON to clean markdown, and writes the results.
func (p *Processor) Process(ctx context.Context, cfg *Config) (*Result, error)  /* v090-result-boundary */ {
	if cfg == nil {
		return nil, core.E("collect.Processor.Process", "config is required", nil)
	}
	ctx, err := activeCollectContext(ctx)
	if err != nil {
		return nil, err
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(p.Name(), "Starting processing")
	}

	dir := p.Dir
	if core.Trim(dir) == "" {
		dir = p.Source
	}
	if dir == "" {
		return &Result{Source: p.Name()}, nil
	}
	result := &Result{Source: p.Name()}
	if emitDryRun(cfg, p.Name(), "[dry-run] Would process files", "Process dry-run complete", result) {
		return result, nil
	}
	if cfg.Output == nil {
		return nil, core.E("collect.Processor.Process", "output medium is required", nil)
	}

	entries, err := cfg.Output.List(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry == nil || entry.IsDir() {
			continue
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}
		p.processEntry(cfg, dir, entry.Name(), result)
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(p.Name(), core.Sprintf("Processed %d files", result.Items), result)
	}
	return result, nil
}

func (p *Processor) processEntry(cfg *Config, dir, name string, result *Result) {
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitProgress(p.Name(), core.Sprintf("Processing %s", name), nil)
	}
	raw, err := cfg.Output.Read(core.JoinPath(dir, name))
	if err != nil {
		p.recordProcessError(cfg, result, core.Sprintf("Failed to read %s: %v", name, err))
		return
	}
	md, err := markdownForFile(name, raw)
	if err != nil {
		p.recordProcessError(cfg, result, core.Sprintf("Failed to convert %s: %v", name, err))
		return
	}
	outName := core.TrimSuffix(name, core.PathExt(name)) + ".md"
	outPath, err := writeResultFile(cfg, p.Name(), outName, md)
	if err != nil {
		p.recordProcessError(cfg, result, core.Sprintf("Failed to write %s: %v", outName, err))
		return
	}
	result.Items++
	result.Files = append(result.Files, outPath)
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitItem(p.Name(), core.Sprintf("Processed %s", name), nil)
	}
}

func (p *Processor) recordProcessError(cfg *Config, result *Result, message string) {
	result.Errors++
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitError(p.Name(), message, nil)
	}
}

func markdownForFile(name, raw string) (string, error)  /* v090-result-boundary */ {
	switch core.Lower(core.PathExt(name)) {
	case ".html", ".htm":
		return HTMLToMarkdown(raw)
	case ".json", ".jsonl":
		return JSONToMarkdown(raw)
	default:
		return raw, nil
	}
}

// HTMLToMarkdown is exported for testing.
func HTMLToMarkdown(content string) (string, error)  /* v090-result-boundary */ {
	if core.Trim(content) == "" {
		return "", nil
	}
	out := content
	replacements := []struct {
		pattern *regexp.Regexp
		value   string
	}{
		{regexp.MustCompile(`(?is)<h1[^>]*>`), "# "},
		{regexp.MustCompile(`(?is)</h1>`), "\n\n"},
		{regexp.MustCompile(`(?is)<h2[^>]*>`), "## "},
		{regexp.MustCompile(`(?is)</h2>`), "\n\n"},
		{regexp.MustCompile(`(?is)<h3[^>]*>`), "### "},
		{regexp.MustCompile(`(?is)</h3>`), "\n\n"},
		{regexp.MustCompile(`(?is)<p[^>]*>`), ""},
		{regexp.MustCompile(`(?is)</p>`), "\n\n"},
		{regexp.MustCompile(`(?is)<br\s*/?>`), "\n"},
		{regexp.MustCompile(`(?is)<strong[^>]*>`), "**"},
		{regexp.MustCompile(`(?is)</strong>`), "**"},
		{regexp.MustCompile(`(?is)<em[^>]*>`), "*"},
		{regexp.MustCompile(`(?is)</em>`), "*"},
	}
	for _, repl := range replacements {
		out = repl.pattern.ReplaceAllString(out, repl.value)
	}
	out = htmlAnchorRe.ReplaceAllStringFunc(out, func(s string) string {
		match := htmlAnchorRe.FindStringSubmatch(s)
		if len(match) != 4 {
			return s
		}
		href := match[1]
		if href == "" {
			href = match[2]
		}
		text := core.Trim(match[3])
		if href == "" {
			return text
		}
		return "[" + text + "](" + href + ")"
	})
	out = htmlTagRe.ReplaceAllString(out, "")
	return core.Trim(out), nil
}

// JSONToMarkdown is exported for testing.
func JSONToMarkdown(content string) (string, error)  /* v090-result-boundary */ {
	if core.Trim(content) == "" {
		return "", nil
	}
	buf := &core.Buffer{}
	buf.WriteString("```json\n")
	var value any
	if r := core.JSONUnmarshal([]byte(content), &value); r.OK {
		if err := encodeJSONValue(buf, value); err != nil {
			return "", err
		}
	} else {
		if ok, err := encodeJSONLines(buf, content); err != nil {
			return "", err
		} else if !ok {
			return "", nil
		}
	}
	buf.WriteString("```\n")
	return core.Trim(buf.String()), nil
}

func encodeJSONValue(buf *core.Buffer, value any) error  /* v090-result-boundary */ {
	r := core.JSONMarshalIndent(value, "", "  ")
	if !r.OK {
		return r.Value.(error)
	}
	buf.Write(r.Value.([]byte))
	buf.WriteByte('\n')
	return nil
}

func encodeJSONLines(buf *core.Buffer, content string) (bool, error)  /* v090-result-boundary */ {
	encoded := false
	for _, line := range core.Split(content, "\n") {
		line = core.Trim(line)
		if line == "" {
			continue
		}
		if encoded {
			buf.WriteString("\n")
		}
		if err := encodeJSONLine(buf, line); err != nil {
			return false, err
		}
		encoded = true
	}
	return encoded, nil
}

func encodeJSONLine(buf *core.Buffer, line string) error  /* v090-result-boundary */ {
	var lineValue any
	if r := core.JSONUnmarshal([]byte(line), &lineValue); !r.OK {
		return r.Value.(error)
	}
	return encodeJSONValue(buf, lineValue)
}
