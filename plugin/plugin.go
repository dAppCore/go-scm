// SPDX-License-Identifier: EUPL-1.2

package plugin

import "context"

type Plugin interface {
	Name() string
	Version() string
	Init(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type BasePlugin struct {
	PluginName    string
	PluginVersion string
}

func (p *BasePlugin) Name() string    { return p.PluginName }
func (p *BasePlugin) Version() string { return p.PluginVersion }
func (p *BasePlugin) Init(context.Context) error  { return nil }
func (p *BasePlugin) Start(context.Context) error { return nil }
func (p *BasePlugin) Stop(context.Context) error  { return nil }
