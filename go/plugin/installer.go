// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	// Note: context.Context is retained as the installer API cancellation contract.
	"context"
	// Note: time is retained for RFC3339 install timestamps in plugin metadata.
	"time"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
)

type Installer struct {
	medium   coreio.Medium
	registry *Registry
}

func NewInstaller(m coreio.Medium, registry *Registry) *Installer {
	return &Installer{medium: m, registry: registry}
}

func ParseSource(source string) (org, repo, version string, err error)  /* v090-result-boundary */ {
	if source == "" {
		return "", "", "", core.E("plugin.ParseSource", "source is required", nil)
	}
	base := source
	if idx := lastIndexAt(source); idx >= 0 {
		base, version = source[:idx], source[idx+1:]
	}
	parts := core.SplitN(base, "/", 2)
	if len(parts) != 2 {
		return "", "", "", core.E("plugin.ParseSource", "expected org/repo or org/repo@version", nil)
	}
	return parts[0], parts[1], version, nil
}

// lastIndexAt returns the byte offset of the last "@" in s, or -1 if absent.
// Equivalent of strings.LastIndex(s, "@") without importing strings.
func lastIndexAt(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '@' {
			return i
		}
	}
	return -1
}

func (i *Installer) Install(ctx context.Context, source string) error  /* v090-result-boundary */ {
	if i == nil {
		return core.E("plugin.Installer.Install", "installer is required", nil)
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	org, repo, version, err := ParseSource(source)
	if err != nil {
		return err
	}
	if i.registry != nil {
		if err := i.registry.Add(&PluginConfig{
			Name:        repo,
			Version:     version,
			Source:      "github:" + org + "/" + repo,
			Enabled:     true,
			InstalledAt: time.Now().UTC().Format(time.RFC3339Nano),
		}); err != nil {
			return err
		}
		if err := i.registry.Save(); err != nil {
			return err
		}
	}
	return nil
}

func (i *Installer) Remove(name string) error  /* v090-result-boundary */ {
	if i == nil {
		return core.E("plugin.Installer.Remove", "installer is required", nil)
	}
	if i.registry != nil {
		if err := i.registry.Remove(name); err != nil {
			return err
		}
		return i.registry.Save()
	}
	return nil
}

func (i *Installer) Update(ctx context.Context, name string) error  /* v090-result-boundary */ {
	if i == nil {
		return core.E("plugin.Installer.Update", "installer is required", nil)
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if i.registry == nil {
		return nil
	}
	cfg, ok := i.registry.Get(name)
	if !ok || cfg == nil {
		return core.E("plugin.Installer.Update", "plugin not found", nil)
	}
	cfg.InstalledAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := i.registry.Add(cfg); err != nil {
		return err
	}
	return i.registry.Save()
}
