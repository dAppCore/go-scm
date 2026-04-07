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
	"reflect"
	"strings"
	"sync"

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
	mu        sync.RWMutex
	index     *marketplace.Index
	installer marketplaceInstaller
	registry  *repos.Registry
	hub       *ws.Hub
	medium    io.Medium
}

type marketplaceInstaller interface {
	Install(context.Context, marketplace.Module) error
	Remove(string) error
	Update(context.Context, string) error
	Installed() ([]marketplace.InstalledModule, error)
}

// compile-time interface checks
var (
	_ provider.Provider    = (*ScmProvider)(nil)
	_ provider.Streamable  = (*ScmProvider)(nil)
	_ provider.Describable = (*ScmProvider)(nil)
	_ provider.Renderable  = (*ScmProvider)(nil)
)

// normaliseInstaller returns nil when inst is a typed nil — e.g. a nil
// *marketplace.Installer wrapped in the interface — so that downstream nil
// checks on p.installer behave correctly.
func normaliseInstaller(inst marketplaceInstaller) marketplaceInstaller {
	if inst == nil {
		return nil
	}
	v := reflect.ValueOf(inst)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	return inst
}

// NewProvider creates an SCM provider backed by the given marketplace index,
// installer, and registry. The WS hub is used to emit real-time events.
// Pass nil for any dependency that is not available.
// Usage: NewProvider(...)
func NewProvider(idx *marketplace.Index, inst marketplaceInstaller, reg *repos.Registry, hub *ws.Hub) *ScmProvider {
	return &ScmProvider{
		index:     idx,
		installer: normaliseInstaller(inst),
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
		"scm.installed.changed",
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
	rg.POST("/marketplace/refresh", p.refreshMarketplace)

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
			Description: "Returns all providers from the marketplace index, optionally filtered by query and category.",
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
			Method:      "POST",
			Path:        "/marketplace/refresh",
			Summary:     "Refresh marketplace index",
			Description: "Reloads an index.json file and replaces the in-memory marketplace catalogue.",
			Tags:        []string{"scm", "marketplace"},
			RequestBody: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"index_path": map[string]any{"type": "string", "description": "Path to an index.json file", "default": "index.json"},
				},
			},
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
	p.mu.RLock()
	idx := p.index
	p.mu.RUnlock()

	if idx == nil {
		c.JSON(http.StatusOK, api.OK([]marketplace.Module{}))
		return
	}

	query := c.Query("q")
	category := c.Query("category")

	modules := idx.Modules
	if category != "" {
		modules = idx.ByCategory(category)
	}
	if query != "" {
		filtered := make([]marketplace.Module, 0, len(modules))
		for _, mod := range modules {
			if moduleMatchesQuery(mod, query) {
				filtered = append(filtered, mod)
			}
		}
		modules = filtered
	}

	if modules == nil {
		modules = []marketplace.Module{}
	}
	c.JSON(http.StatusOK, api.OK(modules))
}

func (p *ScmProvider) getMarketplaceItem(c *gin.Context) {
	p.mu.RLock()
	idx := p.index
	p.mu.RUnlock()

	if idx == nil {
		c.JSON(http.StatusNotFound, api.Fail("not_found", "marketplace index not loaded"))
		return
	}

	code, ok := marketplaceCodeParam(c)
	if !ok {
		return
	}
	mod, ok := idx.Find(code)
	if !ok {
		c.JSON(http.StatusNotFound, api.Fail("not_found", "provider not found in marketplace"))
		return
	}
	c.JSON(http.StatusOK, api.OK(mod))
}

func (p *ScmProvider) installItem(c *gin.Context) {
	p.mu.RLock()
	idx := p.index
	inst := p.installer
	p.mu.RUnlock()

	if idx == nil || inst == nil {
		c.JSON(http.StatusServiceUnavailable, api.Fail("unavailable", "marketplace not configured"))
		return
	}

	code, ok := marketplaceCodeParam(c)
	if !ok {
		return
	}
	mod, ok := idx.Find(code)
	if !ok {
		c.JSON(http.StatusNotFound, api.Fail("not_found", "provider not found in marketplace"))
		return
	}

	if err := inst.Install(c.Request.Context(), mod); err != nil {
		c.JSON(http.StatusInternalServerError, api.Fail("install_failed", err.Error()))
		return
	}

	p.emitEvent("scm.marketplace.installed", map[string]any{
		"code": mod.Code,
		"name": mod.Name,
	})
	p.emitEvent("scm.installed.changed", map[string]any{
		"action": "installed",
		"code":   mod.Code,
		"name":   mod.Name,
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
	p.emitEvent("scm.installed.changed", map[string]any{
		"action": "removed",
		"code":   code,
	})

	c.JSON(http.StatusOK, api.OK(map[string]any{"removed": true, "code": code}))
}

// marketplaceIndexPath is the canonical server-side path for the marketplace
// index file. The refresh endpoint always loads this path — the caller has no
// influence over which file is read.
const marketplaceIndexPath = "index.json"

func (p *ScmProvider) refreshMarketplace(c *gin.Context) {
	// Always use the server-side canonical path. Client-supplied paths are
	// intentionally ignored to prevent path traversal / arbitrary file reads.
	idx, err := marketplace.LoadIndex(p.medium, marketplaceIndexPath)
	if err != nil {
		// Do not expose raw filesystem errors to the caller.
		c.JSON(http.StatusNotFound, api.Fail("index_not_found", "index not found"))
		return
	}

	p.mu.Lock()
	p.index = idx
	p.mu.Unlock()
	p.emitEvent("scm.marketplace.refreshed", map[string]any{
		"index_path": marketplaceIndexPath,
		"modules":    len(idx.Modules),
	})

	c.JSON(http.StatusOK, api.OK(map[string]any{
		"refreshed":  true,
		"index_path": marketplaceIndexPath,
		"modules":    len(idx.Modules),
	}))
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
	if err := p.installer.Update(c.Request.Context(), code); err != nil {
		c.JSON(http.StatusInternalServerError, api.Fail("update_failed", err.Error()))
		return
	}

	p.emitEvent("scm.installed.changed", map[string]any{
		"action": "updated",
		"code":   code,
	})

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

	repoList, err := p.registry.TopologicalOrder()
	if err != nil {
		// Keep the endpoint usable if the registry is malformed.
		repoList = p.registry.List()
	}

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

func moduleMatchesQuery(mod marketplace.Module, query string) bool {
	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(mod.Code), q) ||
		strings.Contains(strings.ToLower(mod.Name), q) ||
		strings.Contains(strings.ToLower(mod.Category), q)
}
