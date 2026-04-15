// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverProvidersReturnsAbsoluteDirs(t *testing.T) {
	root := t.TempDir()
	providerDir := filepath.Join(root, "demo-provider")
	manifestDir := filepath.Join(providerDir, ".core")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(manifestDir, "manifest.yaml"), []byte(`code: demo
name: Demo Provider
version: 1.0.0
namespace: /api/demo
binary: ./bin/demo
`), 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})

	found, err := DiscoverProviders(".")
	if err != nil {
		t.Fatalf("discover providers: %v", err)
	}
	if len(found) != 1 {
		t.Fatalf("expected one provider, got %#v", found)
	}
	wantDir, err := filepath.EvalSymlinks(providerDir)
	if err != nil {
		t.Fatalf("eval symlinks for provider dir: %v", err)
	}
	if got := found[0].Dir; got != wantDir {
		t.Fatalf("expected absolute provider dir %q, got %q", wantDir, got)
	}
}
