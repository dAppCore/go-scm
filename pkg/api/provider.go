// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"dappco.re/go/core/api"
	coreprovider "dappco.re/go/core/api/pkg/provider"
	coreio "dappco.re/go/core/io"
	"dappco.re/go/core/ws"
	"dappco.re/go/scm/marketplace"
	"dappco.re/go/scm/repos"
	"github.com/gin-gonic/gin"
)

type ScmProvider struct {
	index     *marketplace.Index
	installer *marketplace.Installer
	registry  *repos.Registry
	hub       *ws.Hub
	medium    coreio.Medium
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
	rg.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
}

func (p *ScmProvider) Describe() []api.RouteDescription {
	return []api.RouteDescription{{Method: "GET", Path: "/health", Summary: "Health check", StatusCode: 200}}
}

func (p *ScmProvider) Channels() []string { return nil }

func (p *ScmProvider) Element() coreprovider.ElementSpec {
	return coreprovider.ElementSpec{Tag: "core-scm", Source: ""}
}
