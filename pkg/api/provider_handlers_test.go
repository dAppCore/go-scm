// SPDX-Licence-Identifier: EUPL-1.2

package api_test

import (
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	"net/http"
	"net/http/httptest"
	"testing"

	goapi "dappco.re/go/core/api"
	"dappco.re/go/core/scm/marketplace"
	scmapi "dappco.re/go/core/scm/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -- Marketplace: category filter ---------------------------------------------

func TestScmProvider_ListMarketplace_Category_Good(t *testing.T) {
	idx := &marketplace.Index{
		Version: 1,
		Modules: []marketplace.Module{
			{Code: "analytics", Name: "Analytics", Category: "product"},
			{Code: "lint", Name: "Linter", Category: "tool"},
		},
		Categories: []string{"product", "tool"},
	}
	p := scmapi.NewProvider(idx, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace?category=tool", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp goapi.Response[[]marketplace.Module]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "lint", resp.Data[0].Code)
}

// -- Marketplace: nil search results ------------------------------------------

func TestScmProvider_ListMarketplace_SearchNoResults_Good(t *testing.T) {
	idx := &marketplace.Index{
		Version: 1,
		Modules: []marketplace.Module{
			{Code: "analytics", Name: "Analytics", Category: "product"},
		},
	}
	p := scmapi.NewProvider(idx, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace?q=nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// -- GetMarketplaceItem: nil index --------------------------------------------

func TestScmProvider_GetMarketplaceItem_NilIndex_Bad(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace/anything", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// -- Install: nil dependencies ------------------------------------------------

func TestScmProvider_InstallItem_NilDeps_Bad(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/scm/marketplace/test/install", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestScmProvider_InstallItem_NotFound_Bad(t *testing.T) {
	idx := &marketplace.Index{Version: 1}
	p := scmapi.NewProvider(idx, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/scm/marketplace/nonexistent/install", nil)
	r.ServeHTTP(w, req)

	// installer is nil, so it returns unavailable first
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// -- Remove: nil installer ----------------------------------------------------

func TestScmProvider_RemoveItem_NilInstaller_Bad(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/scm/marketplace/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// -- Update: nil installer ----------------------------------------------------

func TestScmProvider_UpdateInstalled_NilInstaller_Bad(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/scm/installed/test/update", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// -- Manifest endpoints: no manifest on disk ----------------------------------

func TestScmProvider_GetManifest_NoFile_Bad(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/manifest", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestScmProvider_GetPermissions_NoFile_Bad(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/manifest/permissions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// -- Verify: bad request ------------------------------------------------------

func TestScmProvider_VerifyManifest_NoBody_Bad(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/scm/manifest/verify", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// -- Sign: bad request --------------------------------------------------------

func TestScmProvider_SignManifest_NoBody_Bad(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/scm/manifest/sign", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// -- Describe: count routes ---------------------------------------------------

func TestScmProvider_Describe_RouteCount_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)
	descs := p.Describe()

	// Verify every route from RegisterRoutes is described
	expectedPaths := []string{
		"/marketplace",
		"/marketplace/:code",
		"/marketplace/:code/install",
		"/manifest",
		"/manifest/verify",
		"/manifest/sign",
		"/manifest/permissions",
		"/installed",
		"/installed/:code/update",
		"/registry",
	}
	paths := make(map[string]bool)
	for _, d := range descs {
		paths[d.Path] = true
	}
	for _, ep := range expectedPaths {
		assert.True(t, paths[ep], "missing description for %s", ep)
	}
}
