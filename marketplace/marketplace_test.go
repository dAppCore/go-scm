// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"testing"

	"dappco.re/go/scm/manifest"
)

func TestBuildIndexFromManifestsCarriesSignKey(t *testing.T) {
	idx := BuildIndexFromManifests([]*manifest.Manifest{
		{
			Code:    "go-io",
			Name:    "Core I/O",
			Layout:  "core",
			SignKey: "ed25519:public-key",
		},
	})

	mod, ok := idx.Find("go-io")
	if !ok {
		t.Fatalf("expected module to be indexed")
	}
	if mod.SignKey != "ed25519:public-key" {
		t.Fatalf("unexpected sign key: %q", mod.SignKey)
	}
}
