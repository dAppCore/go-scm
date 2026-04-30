// SPDX-License-Identifier: EUPL-1.2

// Package pkg wires marketplace package index generation into the scm CLI.
package pkg

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	core "dappco.re/go"
	"dappco.re/go/scm/manifest"
	"dappco.re/go/scm/marketplace"
)

const usage = "usage: scm pkg [--root=DIR] [--dir=DIR] [--dirs=DIR,DIR] [--out=marketplace/index.json] [--base-url=URL] [--org=ORG]"

// Register attaches the pkg command to the parent Core command tree.
//
// The command writes a marketplace index for one package root or a collection
// of package roots. It reads core.json first, then .core/manifest.yaml.
func Register(app *core.Core) core.Result {
	if app == nil {
		return core.Fail(core.E("cmd.pkg.Register", "core app is required", nil))
	}
	return app.Command("pkg", core.Command{Action: run})
}

func run(opts core.Options) core.Result {
	if wantsHelp(opts) {
		core.Print(nil, usage)
		return core.Ok(nil)
	}

	root := option(opts, "root", ".")
	dirs := packageDirs(opts, root)
	idx, err := buildIndex(dirs, opts.String("base-url"), opts.String("org"))
	if err != nil {
		return failed(err)
	}

	outPath := option(opts, "out", filepath.Join(root, "marketplace", "index.json"))
	if err := mkdirParent(outPath); err != nil {
		return failed(err)
	}
	if err := marketplace.WriteIndex(outPath, idx); err != nil {
		return failed(err)
	}

	core.Print(nil, "%s", outPath)
	return core.Ok(nil)
}

func buildIndex(dirs []string, baseURL, org string) (*marketplace.Index, error) {
	var manifests []*manifest.Manifest
	var collectionDirs []string

	for _, dir := range dirs {
		m, err := loadPackageManifest(dir)
		if err == nil {
			manifests = append(manifests, m)
			continue
		}
		collectionDirs = append(collectionDirs, dir)
	}

	idx := marketplace.BuildIndexFromManifests(manifests)
	if len(collectionDirs) == 0 {
		applyRepoDefaults(idx, baseURL, org)
		return idx, nil
	}

	fromDirs, err := (&marketplace.Builder{BaseURL: baseURL, Org: org}).BuildFromDirs(collectionDirs...)
	if err != nil {
		return nil, err
	}
	idx.Modules = append(idx.Modules, fromDirs.Modules...)
	idx.Categories = uniqueCategories(idx.Categories, fromDirs.Categories)
	if idx.Version == 0 {
		idx.Version = 1
	}
	applyRepoDefaults(idx, baseURL, org)
	sortIndex(idx)
	return idx, nil
}

func loadPackageManifest(root string) (*manifest.Manifest, error) {
	if raw, err := os.ReadFile(filepath.Join(root, "core.json")); err == nil {
		cm, err := manifest.ParseCompiled(raw)
		if err != nil {
			return nil, err
		}
		return &cm.Manifest, nil
	}
	raw, err := os.ReadFile(filepath.Join(root, ".core", "manifest.yaml"))
	if err != nil {
		return nil, err
	}
	return manifest.Parse(raw)
}

func packageDirs(opts core.Options, root string) []string {
	if dirs := splitList(opts.String("dirs")); len(dirs) > 0 {
		return dirs
	}
	if dir := strings.TrimSpace(opts.String("dir")); dir != "" {
		return []string{dir}
	}
	if arg := strings.TrimSpace(opts.String("_arg")); arg != "" {
		return []string{arg}
	}
	return []string{root}
}

func uniqueCategories(existing, extra []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(existing)+len(extra))
	for _, category := range append(existing, extra...) {
		category = strings.TrimSpace(category)
		if category == "" {
			continue
		}
		if _, ok := seen[category]; ok {
			continue
		}
		seen[category] = struct{}{}
		out = append(out, category)
	}
	return out
}

func applyRepoDefaults(idx *marketplace.Index, baseURL, org string) {
	if idx == nil || strings.TrimSpace(baseURL) == "" {
		return
	}
	if strings.TrimSpace(org) == "" {
		org = "core"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	for i := range idx.Modules {
		if idx.Modules[i].Repo == "" && idx.Modules[i].Code != "" {
			idx.Modules[i].Repo = baseURL + "/" + org + "/" + idx.Modules[i].Code
		}
	}
}

func sortIndex(idx *marketplace.Index) {
	if idx == nil {
		return
	}
	sort.SliceStable(idx.Modules, func(i, j int) bool {
		return idx.Modules[i].Code < idx.Modules[j].Code
	})
	sort.Strings(idx.Categories)
}

func option(opts core.Options, key, fallback string) string {
	if value := strings.TrimSpace(opts.String(key)); value != "" {
		return value
	}
	return fallback
}

func splitList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}

func mkdirParent(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func wantsHelp(opts core.Options) bool {
	return opts.Bool("help") || opts.Bool("h")
}

func failed(err error) core.Result {
	return core.Fail(err)
}
