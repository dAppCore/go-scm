// SPDX-License-Identifier: EUPL-1.2

package plugin

import (
	// Note: context.Context is retained as the installer API cancellation contract.
	"context"
	// Note: errors.New is retained for stable installer validation errors.
	`errors`
	// Note: strings helpers are retained for parsing plugin source identifiers.
	`strings`
	// Note: time is retained for RFC3339 install timestamps in plugin metadata.
	"time"

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
		return "", "", "", errors.New("plugin.ParseSource: source is required")
	}
	base := source
	if idx := strings.LastIndex(source, "@"); idx >= 0 {
		base, version = source[:idx], source[idx+1:]
	}
	parts := strings.SplitN(base, "/", 2)
	if len(parts) != 2 {
		return "", "", "", errors.New("plugin.ParseSource: expected org/repo or org/repo@version")
	}
	return parts[0], parts[1], version, nil
}

func (i *Installer) Install(ctx context.Context, source string) error  /* v090-result-boundary */ {
	if i == nil {
		return errors.New("plugin.Installer.Install: installer is required")
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
		return errors.New("plugin.Installer.Remove: installer is required")
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
		return errors.New("plugin.Installer.Update: installer is required")
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
		return errors.New("plugin.Installer.Update: plugin not found")
	}
	cfg.InstalledAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := i.registry.Add(cfg); err != nil {
		return err
	}
	return i.registry.Save()
}
