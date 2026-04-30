// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: bytes.Buffer is retained for efficient Markdown assembly in processors.
	"bytes"
	// Note: context.Context is retained as the processor API cancellation contract.
	"context"
	// Note: encoding/json is retained for JSON and JSONL pretty-print processing.
	"encoding/json"
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
func (p *Processor) Process(ctx context.Context, cfg *Config) (*Result, error) {
	if cfg == nil {
		return nil, core.E("collect.Processor.Process", "config is required", nil)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(p.Name(), "[dry-run] Would process files", nil)
			cfg.Dispatcher.EmitComplete(p.Name(), "Process dry-run complete", &Result{Source: p.Name()})
		}
		return &Result{Source: p.Name()}, nil
	}
	if cfg.Output == nil {
		return nil, core.E("collect.Processor.Process", "output medium is required", nil)
	}

	entries, err := cfg.Output.List(dir)
	if err != nil {
		return nil, err
	}
	result := &Result{Source: p.Name()}
	for _, entry := range entries {
		if entry == nil || entry.IsDir() {
			continue
		}
		if ctx != nil {
			if err := ctx.Err(); err != nil {
				return result, err
			}
		}
		name := entry.Name()
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(p.Name(), core.Sprintf("Processing %s", name), nil)
		}
		raw, err := cfg.Output.Read(core.JoinPath(dir, name))
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(p.Name(), core.Sprintf("Failed to read %s: %v", name, err), nil)
			}
			continue
		}
		var md string
		switch core.Lower(core.PathExt(name)) {
		case ".html", ".htm":
			md, err = HTMLToMarkdown(raw)
		case ".json", ".jsonl":
			md, err = JSONToMarkdown(raw)
		default:
			md = raw
		}
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(p.Name(), core.Sprintf("Failed to convert %s: %v", name, err), nil)
			}
			continue
		}
		outName := core.TrimSuffix(name, core.PathExt(name)) + ".md"
		outPath, err := writeResultFile(cfg, p.Name(), outName, md)
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(p.Name(), core.Sprintf("Failed to write %s: %v", outName, err), nil)
			}
			continue
		}
		result.Items++
		result.Files = append(result.Files, outPath)
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitItem(p.Name(), core.Sprintf("Processed %s", name), nil)
		}
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(p.Name(), core.Sprintf("Processed %d files", result.Items), result)
	}
	return result, nil
}

// HTMLToMarkdown is exported for testing.
func HTMLToMarkdown(content string) (string, error) {
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
func JSONToMarkdown(content string) (string, error) {
	if core.Trim(content) == "" {
		return "", nil
	}
	buf := &bytes.Buffer{}
	buf.WriteString("```json\n")
	var value any
	if err := json.Unmarshal([]byte(content), &value); err == nil {
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		if err := enc.Encode(value); err != nil {
			return "", err
		}
	} else {
		lines := core.Split(content, "\n")
		enc := json.NewEncoder(buf)
		enc.SetIndent("", "  ")
		encoded := false
		for _, line := range lines {
			line = core.Trim(line)
			if line == "" {
				continue
			}
			var lineValue any
			if err := json.Unmarshal([]byte(line), &lineValue); err != nil {
				return "", err
			}
			if encoded {
				buf.WriteString("\n")
			}
			if err := enc.Encode(lineValue); err != nil {
				return "", err
			}
			encoded = true
		}
		if !encoded {
			return "", nil
		}
	}
	buf.WriteString("```\n")
	return core.Trim(buf.String()), nil
}
