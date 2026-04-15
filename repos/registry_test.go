// SPDX-License-Identifier: EUPL-1.2

package repos

import (
	"os"
	"path/filepath"
	"testing"

	coreio "dappco.re/go/core/io"
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
