// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
	"dappco.re/go/scm/marketplace"
	"dappco.re/go/scm/repos"
	"github.com/gin-gonic/gin"
)

// doRequest routes a single GET against a freshly-wired provider and
// returns the recorder for status + body assertions.
func doRequest(t *testing.T, provider *ScmProvider, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	provider.RegisterRoutes(router.Group(provider.BasePath()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestProvider_getMarketplaceModule_NilIndex(t *testing.T) {
	rec := doRequest(t, NewProvider(nil, nil, nil, nil), "/scm/marketplace/go-io")
	core.AssertEqual(t, http.StatusNotFound, rec.Code)
}

func TestProvider_getMarketplaceModule_NotFound(t *testing.T) {
	idx := &marketplace.Index{Version: 1, Modules: []marketplace.Module{{Code: "demo"}}}
	rec := doRequest(t, NewProvider(idx, nil, nil, nil), "/scm/marketplace/ghost")
	core.AssertEqual(t, http.StatusNotFound, rec.Code)
}

func TestProvider_listMarketplace_NilIndex(t *testing.T) {
	rec := doRequest(t, NewProvider(nil, nil, nil, nil), "/scm/marketplace")
	core.AssertEqual(t, http.StatusOK, rec.Code)
	core.AssertContains(t, rec.Body.String(), "version")
}

func TestProvider_listRepos_NilRegistry(t *testing.T) {
	rec := doRequest(t, NewProvider(nil, nil, nil, nil), "/scm/repos")
	core.AssertEqual(t, http.StatusOK, rec.Code)
}

func TestProvider_getRepo_NilRegistry(t *testing.T) {
	rec := doRequest(t, NewProvider(nil, nil, nil, nil), "/scm/repos/go-io")
	core.AssertEqual(t, http.StatusNotFound, rec.Code)
}

func TestProvider_getRepo_NotFound(t *testing.T) {
	reg := &repos.Registry{Version: 1, Repos: map[string]*repos.Repo{"other": {Path: "/x"}}}
	rec := doRequest(t, NewProvider(nil, nil, reg, nil), "/scm/repos/go-io")
	core.AssertEqual(t, http.StatusNotFound, rec.Code)
}

func TestProvider_getRepo_Found(t *testing.T) {
	reg := &repos.Registry{Version: 1, Repos: map[string]*repos.Repo{"go-io": {Path: "/workspace/core/go-io"}}}
	rec := doRequest(t, NewProvider(nil, nil, reg, nil), "/scm/repos/go-io")
	core.AssertEqual(t, http.StatusOK, rec.Code)
	core.AssertContains(t, rec.Body.String(), "/workspace/core/go-io")
}

func TestProvider_listInstalledModules_NilInstaller(t *testing.T) {
	rec := doRequest(t, NewProvider(nil, nil, nil, nil), "/scm/modules")
	core.AssertEqual(t, http.StatusOK, rec.Code)
}

func TestProvider_getInstalledModule_NilInstaller(t *testing.T) {
	rec := doRequest(t, NewProvider(nil, nil, nil, nil), "/scm/modules/go-io")
	core.AssertEqual(t, http.StatusNotFound, rec.Code)
}

func TestProvider_getInstalledModule_NotFound(t *testing.T) {
	installer := marketplace.NewInstaller(coreio.NewMockMedium(), "modules")
	rec := doRequest(t, NewProvider(nil, installer, nil, nil), "/scm/modules/ghost")
	core.AssertEqual(t, http.StatusNotFound, rec.Code)
}
