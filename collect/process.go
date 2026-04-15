// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
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
		return nil, errors.New("collect.Processor.Process: config is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if cfg.Output == nil {
		return nil, errors.New("collect.Processor.Process: output medium is required")
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(p.Name(), "Starting processing")
	}

	dir := p.Dir
	if strings.TrimSpace(dir) == "" {
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
		raw, err := cfg.Output.Read(filepath.Join(dir, name))
		if err != nil {
			result.Errors++
			continue
		}
		var md string
		switch strings.ToLower(filepath.Ext(name)) {
		case ".html", ".htm":
			md, err = HTMLToMarkdown(raw)
		case ".json", ".jsonl":
			md, err = JSONToMarkdown(raw)
		default:
			md = raw
		}
		if err != nil {
			result.Errors++
			continue
		}
		outName := strings.TrimSuffix(name, filepath.Ext(name)) + ".md"
		outPath, err := writeResultFile(cfg, p.Name(), outName, md)
		if err != nil {
			result.Errors++
			continue
		}
		result.Items++
		result.Files = append(result.Files, outPath)
	}
	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(p.Name(), fmt.Sprintf("Processed %d files", result.Items), result)
	}
	return result, nil
}

// HTMLToMarkdown is exported for testing.
func HTMLToMarkdown(content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}
	out := content
	replacements := [][2]string{
		{`<h1>`, "# "}, {`</h1>`, "\n\n"},
		{`<h2>`, "## "}, {`</h2>`, "\n\n"},
		{`<h3>`, "### "}, {`</h3>`, "\n\n"},
		{`<p>`, ""}, {`</p>`, "\n\n"},
		{`<br>`, "\n"}, {`<br/>`, "\n"}, {`<br />`, "\n"},
		{`<strong>`, "**"}, {`</strong>`, "**"},
		{`<em>`, "*"}, {`</em>`, "*"},
	}
	for _, repl := range replacements {
		out = strings.ReplaceAll(out, repl[0], repl[1])
		out = strings.ReplaceAll(out, strings.ToUpper(repl[0]), repl[1])
	}
	out = regexp.MustCompile(`(?is)<a[^>]*href="([^"]+)"[^>]*>(.*?)</a>`).ReplaceAllString(out, `[$2]($1)`)
	out = regexp.MustCompile(`(?is)<[^>]+>`).ReplaceAllString(out, "")
	return strings.TrimSpace(out), nil
}

// JSONToMarkdown is exported for testing.
func JSONToMarkdown(content string) (string, error) {
	if strings.TrimSpace(content) == "" {
		return "", nil
	}
	var value any
	if err := json.Unmarshal([]byte(content), &value); err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, "```json")
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		return "", err
	}
	if strings.HasSuffix(buf.String(), "\n") {
		// keep it as-is; the closing fence follows on a new line
	}
	fmt.Fprintln(buf, "```")
	return strings.TrimSpace(buf.String()), nil
}
