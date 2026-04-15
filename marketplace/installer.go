// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	coreio "dappco.re/go/core/io"
	"dappco.re/go/core/io/store"
	"dappco.re/go/scm/manifest"
	"dappco.re/go/scm/internal/ax/jsonx"
)

type InstalledModule struct {
	Code        string               `json:"code"`
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	Repo        string               `json:"repo"`
	EntryPoint  string               `json:"entry_point"`
	Permissions manifest.Permissions `json:"permissions"`
	SignKey     string               `json:"sign_key,omitempty"`
	InstalledAt string               `json:"installed_at"`
}

type Installer struct {
	medium     coreio.Medium
	modulesDir string
	store      *store.KeyValueStore
}

func NewInstaller(m coreio.Medium, modulesDir string, st *store.KeyValueStore) *Installer {
	return &Installer{medium: m, modulesDir: modulesDir, store: st}
}

func (i *Installer) Install(ctx context.Context, mod Module) error {
	if i == nil {
		return errors.New("marketplace.Installer.Install: installer is required")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if i.medium == nil {
		return errors.New("marketplace.Installer.Install: medium is required")
	}
	if mod.Code == "" {
		return errors.New("marketplace.Installer.Install: module code is required")
	}
	entry := InstalledModule{
		Code:        mod.Code,
		Name:        mod.Name,
		Version:     "latest",
		Repo:        mod.Repo,
		EntryPoint:  "core.json",
		InstalledAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	raw, err := jsonx.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return i.medium.Write(filepath.Join(i.modulesDir, mod.Code, "module.json"), string(raw))
}

func (i *Installer) Installed() ([]InstalledModule, error) {
	if i == nil || i.medium == nil {
		return nil, nil
	}
	entries, err := i.medium.List(i.modulesDir)
	if err != nil {
		return nil, err
	}
	var out []InstalledModule
	for _, entry := range entries {
		if entry == nil || !entry.IsDir() {
			continue
		}
		raw, err := i.medium.Read(filepath.Join(i.modulesDir, entry.Name(), "module.json"))
		if err != nil {
			continue
		}
		var mod InstalledModule
		if err := jsonx.Unmarshal([]byte(raw), &mod); err != nil {
			continue
		}
		out = append(out, mod)
	}
	return out, nil
}

func (i *Installer) Remove(code string) error {
	if i == nil || i.medium == nil {
		return errors.New("marketplace.Installer.Remove: installer is required")
	}
	return i.medium.DeleteAll(filepath.Join(i.modulesDir, code))
}

func (i *Installer) Update(ctx context.Context, code string) error {
	if i == nil {
		return errors.New("marketplace.Installer.Update: installer is required")
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}
