// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"crypto/ed25519" // intrinsic
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	coreio "dappco.re/go/io"
	"dappco.re/go/scm/manifest"
	"dappco.re/go/scm/marketplace"
	"dappco.re/go/scm/repos"
	"github.com/gin-gonic/gin"
)

func TestScmProviderRoutesExposeState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	medium := coreio.NewMockMedium()
	installer := marketplace.NewInstaller(medium, "modules")
	mod := signedMarketplaceModule(t, marketplace.Module{
		Code: "go-io",
		Name: "Core I/O",
		Repo: "ssh://git@example.com/core/go-io.git",
	})
	if err := installer.Install(nil, mod); err != nil {
		t.Fatalf("install module: %v", err)
	}

	provider := NewProvider(
		&marketplace.Index{
			Version: 1,
			Modules: []marketplace.Module{mod},
		},
		installer,
		&repos.Registry{
			Version:  1,
			BasePath: "/workspace",
			Repos: map[string]*repos.Repo{
				"go-io": {Path: "/workspace/core/go-io"},
			},
		},
		nil,
	)

	router := gin.New()
	group := router.Group(provider.BasePath())
	provider.RegisterRoutes(group)

	assertRouteOK(t, router, "/scm/health", "ok")
	assertRouteOK(t, router, "/scm/ui", "Source control operations")
	assertRouteOK(t, router, "/scm/marketplace", "go-io")
	assertRouteOK(t, router, "/scm/marketplace/go-io", "Core I/O")
	assertRouteOK(t, router, "/scm/repos", "go-io")
	assertRouteOK(t, router, "/scm/repos/go-io", "/workspace/core/go-io")
	assertRouteOK(t, router, "/scm/modules", "go-io")
	assertRouteOK(t, router, "/scm/modules/go-io", "Core I/O")
}

func TestScmProviderDescribeIncludesReadOnlyRoutes(t *testing.T) {
	provider := NewProvider(nil, nil, nil, nil)
	descs := provider.Describe()

	want := map[string]bool{
		"GET /health":      false,
		"GET /marketplace": false,
		"GET /repos":       false,
		"GET /modules":     false,
	}
	for _, desc := range descs {
		key := desc.Method + " " + desc.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for key, seen := range want {
		if !seen {
			t.Fatalf("expected route description for %s", key)
		}
	}
}

func TestScmProviderMetadataExposesStreamAndElement(t *testing.T) {
	provider := NewProvider(nil, nil, nil, nil)

	if got := provider.Channels(); len(got) != 1 || got[0] != "scm" {
		t.Fatalf("unexpected channels: %#v", got)
	}

	element := provider.Element()
	if element.Tag != "core-scm" {
		t.Fatalf("unexpected element tag: %q", element.Tag)
	}
	if element.Source != "ui/app.js" {
		t.Fatalf("unexpected element source: %q", element.Source)
	}
}

func signedMarketplaceModule(t *testing.T, mod marketplace.Module) marketplace.Module {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	mod.SignKey = base64.StdEncoding.EncodeToString(pub)
	cp := mod
	cp.Sign = ""
	payload, err := json.Marshal(cp)
	if err != nil {
		t.Fatalf("module payload: %v", err)
	}
	sig := &manifest.Manifest{SignKey: mod.SignKey}
	if err := manifest.Sign(sig, payload, priv); err != nil {
		t.Fatalf("sign module: %v", err)
	}
	mod.Sign = sig.Sign
	return mod
}

func assertRouteOK(t *testing.T, router *gin.Engine, path, want string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s: expected 200, got %d with body %s", path, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), want) {
		t.Fatalf("%s: expected body to contain %q, got %s", path, want, rec.Body.String())
	}
}
