// SPDX-License-Identifier: EUPL-1.2

// Package api provides a service provider that wraps go-scm marketplace,
// manifest, and registry functionality as REST endpoints with WebSocket
// event streaming.
package api

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"net/http"
	"net/url"

	"dappco.re/go/core/api"
	"dappco.re/go/core/api/pkg/provider"
	"dappco.re/go/core/io"
	"dappco.re/go/core/scm/agentci"
	"dappco.re/go/core/scm/manifest"
	"dappco.re/go/core/scm/marketplace"
	"dappco.re/go/core/scm/repos"
	"dappco.re/go/core/ws"
	"github.com/gin-gonic/gin"
)

// ScmProvider wraps go-scm marketplace, manifest, and registry operations
// as a service provider. It implements Provider, Streamable, Describable,
// and Renderable.
type ScmProvider struct {
	index     *marketplace.Index
	installer *marketplace.Installer
	registry  *repos.Registry
	hub       *ws.Hub
	medium    io.Medium
}

// compile-time interface checks
var (
	_ provider.Provider    = (*ScmProvider)(nil)
	_ provider.Streamable  = (*ScmProvider)(nil)
	_ provider.Describable = (*ScmProvider)(nil)
	_ provider.Renderable  = (*ScmProvider)(nil)
)

// NewProvider creates an SCM provider backed by the given marketplace index,
// installer, and registry. The WS hub is used to emit real-time events.
// Pass nil for any dependency that is not available.
// Usage: NewProvider(...)
func NewProvider(idx *marketplace.Index, inst *marketplace.Installer, reg *repos.Registry, hub *ws.Hub) *ScmProvider {
	return &ScmProvider{
		index:     idx,
		installer: inst,
		registry:  reg,
		hub:       hub,
		medium:    io.Local,
	}
}

// Name implements api.RouteGroup.
// Usage: Name(...)
func (p *ScmProvider) Name() string { return "scm" }

// BasePath implements api.RouteGroup.
// Usage: BasePath(...)
func (p *ScmProvider) BasePath() string { return "/api/v1/scm" }

// Element implements provider.Renderable.
// Usage: Element(...)
func (p *ScmProvider) Element() provider.ElementSpec {
	return provider.ElementSpec{
		Tag:    "core-scm-panel",
		Source: "/assets/core-scm.js",
	}
}

// Channels implements provider.Streamable.
// Usage: Channels(...)
func (p *ScmProvider) Channels() []string {
	return []string{
		"scm.marketplace.refreshed",
		"scm.marketplace.installed",
		"scm.marketplace.removed",
		"scm.manifest.verified",
		"scm.registry.changed",
	}
}

// RegisterRoutes implements api.RouteGroup.
// Usage: RegisterRoutes(...)
func (p *ScmProvider) RegisterRoutes(rg *gin.RouterGroup) {
	// Marketplace
	rg.GET("/marketplace", p.listMarketplace)
	rg.GET("/marketplace/:code", p.getMarketplaceItem)
	rg.POST("/marketplace/:code/install", p.installItem)
	rg.DELETE("/marketplace/:code", p.removeItem)

	// Manifest
	rg.GET("/manifest", p.getManifest)
	rg.POST("/manifest/verify", p.verifyManifest)
	rg.POST("/manifest/sign", p.signManifest)
	rg.GET("/manifest/permissions", p.getPermissions)

	// Installed
	rg.GET("/installed", p.listInstalled)
	rg.POST("/installed/:code/update", p.updateInstalled)

	// Registry
	rg.GET("/registry", p.listRegistry)
}

// Describe implements api.DescribableGroup.
// Usage: Describe(...)
func (p *ScmProvider) Describe() []api.RouteDescription {
	return []api.RouteDescription{
		{
			Method:      "GET",
			Path:        "/marketplace",
			Summary:     "List available providers",
			Description: "Returns all providers from the marketplace index, optionally filtered by query or category.",
			Tags:        []string{"scm", "marketplace"},
		},
		{
			Method:      "GET",
			Path:        "/marketplace/:code",
			Summary:     "Get provider details",
			Description: "Returns a single provider entry from the marketplace by its code.",
			Tags:        []string{"scm", "marketplace"},
		},
		{
			Method:      "POST",
			Path:        "/marketplace/:code/install",
			Summary:     "Install a provider",
			Description: "Clones the provider repository, verifies its manifest signature, and registers it.",
			Tags:        []string{"scm", "marketplace"},
		},
		{
			Method:      "DELETE",
			Path:        "/marketplace/:code",
			Summary:     "Remove an installed provider",
			Description: "Uninstalls a provider by removing its files and store entry.",
			Tags:        []string{"scm", "marketplace"},
		},
		{
			Method:      "GET",
			Path:        "/manifest",
			Summary:     "Read manifest",
			Description: "Reads and parses the .core/manifest.yaml from the current directory.",
			Tags:        []string{"scm", "manifest"},
		},
		{
			Method:      "POST",
			Path:        "/manifest/verify",
			Summary:     "Verify manifest signature",
			Description: "Verifies the Ed25519 signature of the manifest using the provided public key.",
			Tags:        []string{"scm", "manifest"},
			RequestBody: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"public_key": map[string]any{"type": "string", "description": "Hex-encoded Ed25519 public key"},
				},
			},
		},
		{
			Method:      "POST",
			Path:        "/manifest/sign",
			Summary:     "Sign manifest",
			Description: "Signs the manifest with the provided Ed25519 private key.",
			Tags:        []string{"scm", "manifest"},
			RequestBody: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"private_key": map[string]any{"type": "string", "description": "Hex-encoded Ed25519 private key"},
				},
			},
		},
		{
			Method:      "GET",
			Path:        "/manifest/permissions",
			Summary:     "List manifest permissions",
			Description: "Returns the declared permissions from the current manifest.",
			Tags:        []string{"scm", "manifest"},
		},
		{
			Method:      "GET",
			Path:        "/installed",
			Summary:     "List installed providers",
			Description: "Returns all installed provider metadata.",
			Tags:        []string{"scm", "installed"},
		},
		{
			Method:      "POST",
			Path:        "/installed/:code/update",
			Summary:     "Update an installed provider",
			Description: "Pulls latest changes from git and re-verifies the manifest.",
			Tags:        []string{"scm", "installed"},
		},
		{
			Method:      "GET",
			Path:        "/registry",
			Summary:     "List registry repos",
			Description: "Returns all repositories from the repos.yaml registry.",
			Tags:        []string{"scm", "registry"},
		},
	}
}

// -- Marketplace Handlers -----------------------------------------------------

func (p *ScmProvider) listMarketplace(c *gin.Context) {
	if p.index == nil {
		c.JSON(http.StatusOK, api.OK([]marketplace.Module{}))
		return
	}

	query := c.Query("q")
	category := c.Query("category")

	var modules []marketplace.Module
	switch {
	case query != "":
		modules = p.index.Search(query)
	case category != "":
		modules = p.index.ByCategory(category)
	default:
		modules = p.index.Modules
	}

	if modules == nil {
		modules = []marketplace.Module{}
	}
	c.JSON(http.StatusOK, api.OK(modules))
}

func (p *ScmProvider) getMarketplaceItem(c *gin.Context) {
	if p.index == nil {
		c.JSON(http.StatusNotFound, api.Fail("not_found", "marketplace index not loaded"))
		return
	}

	code, ok := marketplaceCodeParam(c)
	if !ok {
		return
	}
	mod, ok := p.index.Find(code)
	if !ok {
		c.JSON(http.StatusNotFound, api.Fail("not_found", "provider not found in marketplace"))
		return
	}
	c.JSON(http.StatusOK, api.OK(mod))
}

func (p *ScmProvider) installItem(c *gin.Context) {
	if p.index == nil || p.installer == nil {
		c.JSON(http.StatusServiceUnavailable, api.Fail("unavailable", "marketplace not configured"))
		return
	}

	code, ok := marketplaceCodeParam(c)
	if !ok {
		return
	}
	mod, ok := p.index.Find(code)
	if !ok {
		c.JSON(http.StatusNotFound, api.Fail("not_found", "provider not found in marketplace"))
		return
	}

	if err := p.installer.Install(context.Background(), mod); err != nil {
		c.JSON(http.StatusInternalServerError, api.Fail("install_failed", err.Error()))
		return
	}

	p.emitEvent("scm.marketplace.installed", map[string]any{
		"code": mod.Code,
		"name": mod.Name,
	})

	c.JSON(http.StatusOK, api.OK(map[string]any{"installed": true, "code": mod.Code}))
}

func (p *ScmProvider) removeItem(c *gin.Context) {
	if p.installer == nil {
		c.JSON(http.StatusServiceUnavailable, api.Fail("unavailable", "installer not configured"))
		return
	}

	code, ok := marketplaceCodeParam(c)
	if !ok {
		return
	}
	if err := p.installer.Remove(code); err != nil {
		c.JSON(http.StatusInternalServerError, api.Fail("remove_failed", err.Error()))
		return
	}

	p.emitEvent("scm.marketplace.removed", map[string]any{"code": code})

	c.JSON(http.StatusOK, api.OK(map[string]any{"removed": true, "code": code}))
}

// -- Manifest Handlers --------------------------------------------------------

func (p *ScmProvider) getManifest(c *gin.Context) {
	m, err := manifest.Load(p.medium, ".")
	if err != nil {
		c.JSON(http.StatusNotFound, api.Fail("manifest_not_found", err.Error()))
		return
	}
	c.JSON(http.StatusOK, api.OK(m))
}

type verifyRequest struct {
	PublicKey string `json:"public_key" binding:"required"`
}

func (p *ScmProvider) verifyManifest(c *gin.Context) {
	var req verifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Fail("invalid_request", "public_key is required"))
		return
	}

	m, err := manifest.Load(p.medium, ".")
	if err != nil {
		c.JSON(http.StatusNotFound, api.Fail("manifest_not_found", err.Error()))
		return
	}

	pubBytes, err := hex.DecodeString(req.PublicKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Fail("invalid_key", "public key must be hex-encoded"))
		return
	}

	valid, err := manifest.Verify(m, ed25519.PublicKey(pubBytes))
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, api.Fail("verify_failed", err.Error()))
		return
	}

	p.emitEvent("scm.manifest.verified", map[string]any{
		"code":  m.Code,
		"valid": valid,
	})

	c.JSON(http.StatusOK, api.OK(map[string]any{"valid": valid, "code": m.Code}))
}

type signRequest struct {
	PrivateKey string `json:"private_key" binding:"required"`
}

func (p *ScmProvider) signManifest(c *gin.Context) {
	var req signRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.Fail("invalid_request", "private_key is required"))
		return
	}

	m, err := manifest.Load(p.medium, ".")
	if err != nil {
		c.JSON(http.StatusNotFound, api.Fail("manifest_not_found", err.Error()))
		return
	}

	privBytes, err := hex.DecodeString(req.PrivateKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Fail("invalid_key", "private key must be hex-encoded"))
		return
	}

	if err := manifest.Sign(m, ed25519.PrivateKey(privBytes)); err != nil {
		c.JSON(http.StatusInternalServerError, api.Fail("sign_failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, api.OK(map[string]any{"signed": true, "code": m.Code, "signature": m.Sign}))
}

func (p *ScmProvider) getPermissions(c *gin.Context) {
	m, err := manifest.Load(p.medium, ".")
	if err != nil {
		c.JSON(http.StatusNotFound, api.Fail("manifest_not_found", err.Error()))
		return
	}
	c.JSON(http.StatusOK, api.OK(m.Permissions))
}

// -- Installed Handlers -------------------------------------------------------

func (p *ScmProvider) listInstalled(c *gin.Context) {
	if p.installer == nil {
		c.JSON(http.StatusOK, api.OK([]marketplace.InstalledModule{}))
		return
	}

	installed, err := p.installer.Installed()
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Fail("list_failed", err.Error()))
		return
	}
	if installed == nil {
		installed = []marketplace.InstalledModule{}
	}
	c.JSON(http.StatusOK, api.OK(installed))
}

func (p *ScmProvider) updateInstalled(c *gin.Context) {
	if p.installer == nil {
		c.JSON(http.StatusServiceUnavailable, api.Fail("unavailable", "installer not configured"))
		return
	}

	code, ok := marketplaceCodeParam(c)
	if !ok {
		return
	}
	if err := p.installer.Update(context.Background(), code); err != nil {
		c.JSON(http.StatusInternalServerError, api.Fail("update_failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, api.OK(map[string]any{"updated": true, "code": code}))
}

// -- Registry Handlers --------------------------------------------------------

// repoSummary is a JSON-friendly representation of a registry repo.
type repoSummary struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty"`
	Path        string   `json:"path,omitempty"`
	Exists      bool     `json:"exists"`
}

func (p *ScmProvider) listRegistry(c *gin.Context) {
	if p.registry == nil {
		c.JSON(http.StatusOK, api.OK([]repoSummary{}))
		return
	}

	repoList := p.registry.List()
	summaries := make([]repoSummary, 0, len(repoList))
	for _, r := range repoList {
		summaries = append(summaries, repoSummary{
			Name:        r.Name,
			Type:        r.Type,
			Description: r.Description,
			DependsOn:   r.DependsOn,
			Path:        r.Path,
			Exists:      r.Exists(),
		})
	}

	c.JSON(http.StatusOK, api.OK(summaries))
}

// -- Internal Helpers ---------------------------------------------------------

// emitEvent sends a WS event if the hub is available.
func (p *ScmProvider) emitEvent(channel string, data any) {
	if p.hub == nil {
		return
	}
	_ = p.hub.SendToChannel(channel, ws.Message{
		Type: ws.TypeEvent,
		Data: data,
	})
}

func marketplaceCodeParam(c *gin.Context) (string, bool) {
	code, err := normaliseMarketplaceCode(c.Param("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Fail("invalid_code", "invalid marketplace code"))
		return "", false
	}
	return code, true
}

func normaliseMarketplaceCode(raw string) (string, error) {
	decoded, err := url.PathUnescape(raw)
	if err != nil {
		return "", err
	}

	return agentci.ValidatePathElement(decoded)
}
