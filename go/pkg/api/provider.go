// SPDX-License-Identifier: EUPL-1.2

package api

import (
	// Note: AX-6 — Provider routes expose HTTP status codes at the Gin boundary.
	"net/http"

	core "dappco.re/go"
	"dappco.re/go/scm/marketplace"
	"dappco.re/go/scm/repos"
	"dappco.re/go/ws"
	"github.com/gin-gonic/gin"
)

const (
	sonarProviderModuleNotFound = "module not found"
)

type ScmProvider struct {
	index     *marketplace.Index
	installer *marketplace.Installer
	registry  *repos.Registry
	hub       *ws.Hub
}

type elementSpec struct {
	Tag    string `json:"tag"`
	Source string `json:"source"`
}

func NewProvider(idx *marketplace.Index, inst *marketplace.Installer, reg *repos.Registry, hub *ws.Hub) *ScmProvider {
	return &ScmProvider{index: idx, installer: inst, registry: reg, hub: hub}
}

func (p *ScmProvider) Name() string     { return "scm" }
func (p *ScmProvider) BasePath() string { return "/scm" }

func (p *ScmProvider) RegisterRoutes(rg *gin.RouterGroup) {
	if rg == nil {
		return
	}
	rg.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	rg.GET("/ui", p.serveUI)
	rg.StaticFS("/ui/assets", embeddedUI)
	rg.GET("/marketplace", p.listMarketplace)
	rg.GET("/marketplace/:code", p.getMarketplaceModule)
	rg.GET("/repos", p.listRepos)
	rg.GET("/repos/:name", p.getRepo)
	rg.GET("/modules", p.listInstalledModules)
	rg.GET("/modules/:code", p.getInstalledModule)
}

func (p *ScmProvider) Channels() []string {
	return []string{"scm"}
}

func (p *ScmProvider) Element() elementSpec {
	return elementSpec{Tag: "core-scm", Source: "ui/app.js"}
}

func (p *ScmProvider) serveUI(c *gin.Context) {
	if c == nil {
		return
	}
	if len(embeddedIndexHTML) == 0 {
		c.Status(http.StatusNotFound)
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", embeddedIndexHTML)
}

func (p *ScmProvider) listMarketplace(c *gin.Context) {
	if c == nil {
		return
	}
	if p == nil || p.index == nil {
		c.JSON(http.StatusOK, marketplace.Index{Version: 1, Modules: []marketplace.Module{}})
		return
	}
	c.JSON(http.StatusOK, p.index)
}

func (p *ScmProvider) getMarketplaceModule(c *gin.Context) {
	if c == nil {
		return
	}
	if p == nil || p.index == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": sonarProviderModuleNotFound})
		return
	}
	code := core.Trim(c.Param("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}
	mod, ok := p.index.Find(code)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": sonarProviderModuleNotFound})
		return
	}
	c.JSON(http.StatusOK, mod)
}

func (p *ScmProvider) listRepos(c *gin.Context) {
	if c == nil {
		return
	}
	if p == nil || p.registry == nil {
		c.JSON(http.StatusOK, repos.Registry{Version: 1, Repos: map[string]*repos.Repo{}})
		return
	}
	c.JSON(http.StatusOK, p.registry)
}

func (p *ScmProvider) getRepo(c *gin.Context) {
	if c == nil {
		return
	}
	if p == nil || p.registry == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "repo not found"})
		return
	}
	name := core.Trim(c.Param("name"))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	repo, ok := p.registry.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "repo not found"})
		return
	}
	c.JSON(http.StatusOK, repo)
}

func (p *ScmProvider) listInstalledModules(c *gin.Context) {
	if c == nil {
		return
	}
	if p == nil || p.installer == nil {
		c.JSON(http.StatusOK, []marketplace.InstalledModule{})
		return
	}
	modules, err := p.installer.Installed()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, modules)
}

func (p *ScmProvider) getInstalledModule(c *gin.Context) {
	if c == nil {
		return
	}
	code := core.Trim(c.Param("code"))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}
	if p == nil || p.installer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": sonarProviderModuleNotFound})
		return
	}
	modules, err := p.installer.Installed()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, mod := range modules {
		if core.Lower(mod.Code) == core.Lower(code) {
			c.JSON(http.StatusOK, mod)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": sonarProviderModuleNotFound})
}
