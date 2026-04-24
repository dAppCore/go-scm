// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"os"
	"path/filepath"
	"testing"

	coreio "dappco.re/go/io"
)

func TestFindRegistryHonorsCORE_REPOS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom-repos.yaml")
	if err := os.WriteFile(path, []byte("version: 1\nrepos: {}\n"), 0o600); err != nil {
		t.Fatalf("write registry: %v", err)
	}
	t.Setenv("CORE_REPOS", path)

	got, err := FindRegistry(coreio.Local)
	if err != nil {
		t.Fatalf("FindRegistry: %v", err)
	}
	if got != path {
		t.Fatalf("expected %q, got %q", path, got)
	}
}

func TestLoadRegistryPreservesMediumForSave(t *testing.T) {
	medium := coreio.NewMockMedium()
	path := filepath.Join(t.TempDir(), "repos.yaml")
	if err := medium.Write(path, "version: 1\nrepos:\n  demo:\n    path: demo\n"); err != nil {
		t.Fatalf("seed registry: %v", err)
	}

	reg, err := LoadRegistry(medium, path)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}
	if reg == nil {
		t.Fatal("expected registry")
	}
	reg.Defaults.Branch = "dev"

	if err := reg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, ok := medium.Files[path]
	if !ok {
		t.Fatalf("expected registry save to use the original medium")
	}
	if raw == "" {
		t.Fatalf("expected saved registry content")
	}
}
