// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
	"dappco.re/go/scm/manifest"
	"dappco.re/go/scm/marketplace"
	"dappco.re/go/scm/repos"
	"github.com/gin-gonic/gin"
)

const (
	sonarProviderTestCoreIO  = "Core I/O"
	sonarProviderTestCoreScm = "core-scm"
	sonarProviderTestUiAppJs = "ui/app.js"
)

func TestScmProviderRoutesExposeState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	medium := coreio.NewMockMedium()
	installer := marketplace.NewInstaller(medium, "modules")
	mod := signedMarketplaceModule(t, marketplace.Module{
		Code: "go-io",
		Name: sonarProviderTestCoreIO,
		Repo: "ssh://git@example.com/core/go-io.git",
	})
	var ctx context.Context
	if err := installer.Install(ctx, mod); err != nil {
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
	assertRouteOK(t, router, "/scm/marketplace/go-io", sonarProviderTestCoreIO)
	assertRouteOK(t, router, "/scm/repos", "go-io")
	assertRouteOK(t, router, "/scm/repos/go-io", "/workspace/core/go-io")
	assertRouteOK(t, router, "/scm/modules", "go-io")
	assertRouteOK(t, router, "/scm/modules/go-io", sonarProviderTestCoreIO)
}

func TestScmProviderMetadataExposesStreamAndElement(t *testing.T) {
	provider := NewProvider(nil, nil, nil, nil)

	if got := provider.Channels(); len(got) != 1 || got[0] != "scm" {
		t.Fatalf("unexpected channels: %#v", got)
	}

	element := provider.Element()
	if element.Tag != sonarProviderTestCoreScm {
		t.Fatalf("unexpected element tag: %q", element.Tag)
	}
	if element.Source != sonarProviderTestUiAppJs {
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
	r := core.JSONMarshal(cp)
	if !r.OK {
		t.Fatalf("module payload: %v", r.Value)
	}
	payload, _ := r.Value.([]byte)
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
	if !core.Contains(rec.Body.String(), want) {
		t.Fatalf("%s: expected body to contain %q, got %s", path, want, rec.Body.String())
	}
}

func TestProvider_NewProvider_Good(t *core.T) {
	provider := NewProvider(&marketplace.Index{Version: 1}, nil, &repos.Registry{Version: 1}, nil)
	core.AssertNotNil(t, provider.index)
	core.AssertNotNil(t, provider.registry)
}

func TestProvider_NewProvider_Bad(t *core.T) {
	provider := NewProvider(nil, nil, nil, nil)
	core.AssertNil(t, provider.index)
	core.AssertNil(t, provider.registry)
}

func TestProvider_NewProvider_Ugly(t *core.T) {
	hubProvider := NewProvider(nil, nil, nil, nil)
	core.AssertNil(t, hubProvider.hub)
	core.AssertEqual(t, "scm", hubProvider.Name())
}

func TestProvider_ScmProvider_Name_Good(t *core.T) {
	provider := NewProvider(nil, nil, nil, nil)
	got := provider.Name()
	core.AssertEqual(t, "scm", got)
}

func TestProvider_ScmProvider_Name_Bad(t *core.T) {
	provider := &ScmProvider{}
	got := provider.Name()
	core.AssertEqual(t, "scm", got)
}

func TestProvider_ScmProvider_Name_Ugly(t *core.T) {
	var provider *ScmProvider
	got := provider.Name()
	core.AssertEqual(t, "scm", got)
}

func TestProvider_ScmProvider_BasePath_Good(t *core.T) {
	provider := NewProvider(nil, nil, nil, nil)
	got := provider.BasePath()
	core.AssertEqual(t, "/scm", got)
}

func TestProvider_ScmProvider_BasePath_Bad(t *core.T) {
	provider := &ScmProvider{}
	got := provider.BasePath()
	core.AssertEqual(t, "/scm", got)
}

func TestProvider_ScmProvider_BasePath_Ugly(t *core.T) {
	var provider *ScmProvider
	got := provider.BasePath()
	core.AssertEqual(t, "/scm", got)
}

func TestProvider_ScmProvider_RegisterRoutes_Good(t *core.T) {
	gin.SetMode(gin.TestMode)
	provider := NewProvider(nil, nil, nil, nil)
	router := gin.New()
	provider.RegisterRoutes(router.Group(provider.BasePath()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/scm/health", nil))
	core.AssertEqual(t, http.StatusOK, rec.Code)
}

func TestProvider_ScmProvider_RegisterRoutes_Bad(t *core.T) {
	provider := NewProvider(nil, nil, nil, nil)
	core.AssertNotPanics(t, func() { provider.RegisterRoutes(nil) })
	core.AssertEqual(t, "scm", provider.Name())
}

func TestProvider_ScmProvider_RegisterRoutes_Ugly(t *core.T) {
	gin.SetMode(gin.TestMode)
	var provider *ScmProvider
	router := gin.New()
	core.AssertNotPanics(t, func() { provider.RegisterRoutes(router.Group("/scm")) })
	core.AssertEqual(t, "scm", provider.Name())
}

func TestProvider_ScmProvider_Channels_Good(t *core.T) {
	provider := NewProvider(nil, nil, nil, nil)
	got := provider.Channels()
	core.AssertEqual(t, []string{"scm"}, got)
}

func TestProvider_ScmProvider_Channels_Bad(t *core.T) {
	provider := &ScmProvider{}
	got := provider.Channels()
	core.AssertLen(t, got, 1)
}

func TestProvider_ScmProvider_Channels_Ugly(t *core.T) {
	var provider *ScmProvider
	got := provider.Channels()
	core.AssertEqual(t, "scm", got[0])
}

func TestProvider_ScmProvider_Element_Good(t *core.T) {
	provider := NewProvider(nil, nil, nil, nil)
	got := provider.Element()
	core.AssertEqual(t, sonarProviderTestCoreScm, got.Tag)
	core.AssertEqual(t, sonarProviderTestUiAppJs, got.Source)
}

func TestProvider_ScmProvider_Element_Bad(t *core.T) {
	provider := &ScmProvider{}
	got := provider.Element()
	core.AssertEqual(t, sonarProviderTestCoreScm, got.Tag)
}

func TestProvider_ScmProvider_Element_Ugly(t *core.T) {
	var provider *ScmProvider
	got := provider.Element()
	core.AssertEqual(t, sonarProviderTestUiAppJs, got.Source)
}
