// SPDX-License-Identifier: EUPL-1.2

package api_test

import (
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	"net/http"
	"net/http/httptest"
	"testing"

	goapi "dappco.re/go/core/api"
	"dappco.re/go/core/scm/marketplace"
	scmapi "dappco.re/go/core/scm/pkg/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// -- Provider Identity --------------------------------------------------------

func TestScmProvider_Name_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)
	assert.Equal(t, "scm", p.Name())
}

func TestScmProvider_BasePath_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)
	assert.Equal(t, "/api/v1/scm", p.BasePath())
}

func TestScmProvider_Channels_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)
	channels := p.Channels()
	assert.Contains(t, channels, "scm.marketplace.refreshed")
	assert.Contains(t, channels, "scm.marketplace.installed")
	assert.Contains(t, channels, "scm.marketplace.removed")
	assert.Contains(t, channels, "scm.manifest.verified")
	assert.Contains(t, channels, "scm.registry.changed")
}

func TestScmProvider_Element_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)
	el := p.Element()
	assert.Equal(t, "core-scm-panel", el.Tag)
	assert.Equal(t, "/assets/core-scm.js", el.Source)
}

func TestScmProvider_Describe_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)
	descs := p.Describe()
	assert.GreaterOrEqual(t, len(descs), 11)

	for _, d := range descs {
		assert.NotEmpty(t, d.Method)
		assert.NotEmpty(t, d.Path)
		assert.NotEmpty(t, d.Summary)
		assert.NotEmpty(t, d.Tags)
	}
}

// -- Marketplace Endpoints ----------------------------------------------------

func TestScmProvider_ListMarketplace_Good(t *testing.T) {
	idx := &marketplace.Index{
		Version: 1,
		Modules: []marketplace.Module{
			{Code: "analytics", Name: "Analytics", Category: "product"},
			{Code: "bio", Name: "Bio Links", Category: "product"},
		},
		Categories: []string{"product"},
	}
	p := scmapi.NewProvider(idx, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp goapi.Response[[]marketplace.Module]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data, 2)
}

func TestScmProvider_ListMarketplace_Search_Good(t *testing.T) {
	idx := &marketplace.Index{
		Version: 1,
		Modules: []marketplace.Module{
			{Code: "analytics", Name: "Analytics", Category: "product"},
			{Code: "bio", Name: "Bio Links", Category: "product"},
		},
	}
	p := scmapi.NewProvider(idx, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace?q=bio", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp goapi.Response[[]marketplace.Module]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "bio", resp.Data[0].Code)
}

func TestScmProvider_ListMarketplace_NilIndex_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp goapi.Response[[]marketplace.Module]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Empty(t, resp.Data)
}

func TestScmProvider_GetMarketplaceItem_Good(t *testing.T) {
	idx := &marketplace.Index{
		Version: 1,
		Modules: []marketplace.Module{
			{Code: "analytics", Name: "Analytics", Category: "product"},
		},
	}
	p := scmapi.NewProvider(idx, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace/analytics", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp goapi.Response[marketplace.Module]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "analytics", resp.Data.Code)
}

func TestScmProvider_GetMarketplaceItem_Bad(t *testing.T) {
	idx := &marketplace.Index{Version: 1}
	p := scmapi.NewProvider(idx, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestScmProvider_GetMarketplaceItem_Bad_PathTraversal(t *testing.T) {
	idx := &marketplace.Index{Version: 1}
	p := scmapi.NewProvider(idx, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/marketplace/%2e%2e", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// -- Installed Endpoints ------------------------------------------------------

func TestScmProvider_ListInstalled_NilInstaller_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/installed", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp goapi.Response[[]marketplace.InstalledModule]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Empty(t, resp.Data)
}

// -- Registry Endpoints -------------------------------------------------------

func TestScmProvider_ListRegistry_NilRegistry_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	r := setupRouter(p)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scm/registry", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp goapi.Response[[]any]
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Empty(t, resp.Data)
}

// -- Route Registration -------------------------------------------------------

func TestScmProvider_RegistersAsRouteGroup_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	engine, err := goapi.New()
	require.NoError(t, err)

	engine.Register(p)
	assert.Len(t, engine.Groups(), 1)
	assert.Equal(t, "scm", engine.Groups()[0].Name())
}

func TestScmProvider_Channels_RegisterAsStreamGroup_Good(t *testing.T) {
	p := scmapi.NewProvider(nil, nil, nil, nil)

	engine, err := goapi.New()
	require.NoError(t, err)

	engine.Register(p)

	channels := engine.Channels()
	assert.Contains(t, channels, "scm.marketplace.refreshed")
}

// -- Test helpers -------------------------------------------------------------

func setupRouter(p *scmapi.ScmProvider) *gin.Engine {
	r := gin.New()
	rg := r.Group(p.BasePath())
	p.RegisterRoutes(rg)
	return r
}
