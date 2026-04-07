// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	json "dappco.re/go/core/scm/internal/ax/jsonx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"encoding/hex"
	exec "golang.org/x/sys/execabs"
	"time"

	"dappco.re/go/core/io"
	"dappco.re/go/core/io/store"
	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/scm/agentci"
	"dappco.re/go/core/scm/manifest"
)

const storeGroup = "_modules"

// Installer handles module installation from Git repos.
type Installer struct {
	medium     io.Medium
	modulesDir string
	store      *store.KeyValueStore
}

// NewInstaller creates a new module installer.
// Usage: NewInstaller(...)
func NewInstaller(m io.Medium, modulesDir string, st *store.KeyValueStore) *Installer {
	return &Installer{
		medium:     m,
		modulesDir: modulesDir,
		store:      st,
	}
}

// InstalledModule holds stored metadata about an installed module.
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

// Install clones a module repo, verifies its manifest signature, and registers it.
// Usage: Install(...)
func (i *Installer) Install(ctx context.Context, mod Module) error {
	safeCode, dest, err := i.resolveModulePath(mod.Code)
	if err != nil {
		return coreerr.E("marketplace.Installer.Install", "invalid module code", err)
	}

	// Check if already installed
	if _, err := i.store.Get(storeGroup, safeCode); err == nil {
		return coreerr.E("marketplace.Installer.Install", "module already installed: "+safeCode, nil)
	}

	if err := i.medium.EnsureDir(i.modulesDir); err != nil {
		return coreerr.E("marketplace.Installer.Install", "mkdir", err)
	}
	if err := gitClone(ctx, mod.Repo, dest); err != nil {
		return coreerr.E("marketplace.Installer.Install", "clone "+mod.Repo, err)
	}

	// On any error after clone, clean up the directory
	cleanup := true
	defer func() {
		if cleanup {
			_ = i.medium.DeleteAll(dest)
		}
	}()

	medium, err := io.NewSandboxed(dest)
	if err != nil {
		return coreerr.E("marketplace.Installer.Install", "medium", err)
	}

	m, err := loadManifest(medium, mod.SignKey)
	if err != nil {
		return err
	}

	entryPoint := filepath.Join(dest, "main.ts")
	installed := InstalledModule{
		Code:        safeCode,
		Name:        m.Name,
		Version:     m.Version,
		Repo:        mod.Repo,
		EntryPoint:  entryPoint,
		Permissions: m.Permissions,
		SignKey:     mod.SignKey,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(installed)
	if err != nil {
		return coreerr.E("marketplace.Installer.Install", "marshal", err)
	}

	if err := i.store.Set(storeGroup, safeCode, string(data)); err != nil {
		return coreerr.E("marketplace.Installer.Install", "store", err)
	}

	cleanup = false
	return nil
}

// Remove uninstalls a module by deleting its files and store entry.
// Usage: Remove(...)
func (i *Installer) Remove(code string) error {
	safeCode, dest, err := i.resolveModulePath(code)
	if err != nil {
		return coreerr.E("marketplace.Installer.Remove", "invalid module code", err)
	}

	if _, err := i.store.Get(storeGroup, safeCode); err != nil {
		return coreerr.E("marketplace.Installer.Remove", "module not installed: "+safeCode, nil)
	}

	if err := i.medium.DeleteAll(dest); err != nil {
		return coreerr.E("marketplace.Installer.Remove", "delete module files", err)
	}

	return i.store.Delete(storeGroup, safeCode)
}

// Update pulls latest changes and re-verifies the manifest.
// Usage: Update(...)
func (i *Installer) Update(ctx context.Context, code string) error {
	safeCode, dest, err := i.resolveModulePath(code)
	if err != nil {
		return coreerr.E("marketplace.Installer.Update", "invalid module code", err)
	}

	raw, err := i.store.Get(storeGroup, safeCode)
	if err != nil {
		return coreerr.E("marketplace.Installer.Update", "module not installed: "+safeCode, nil)
	}

	var installed InstalledModule
	if err := json.Unmarshal([]byte(raw), &installed); err != nil {
		return coreerr.E("marketplace.Installer.Update", "unmarshal", err)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", dest, "pull", "--ff-only")
	if output, err := cmd.CombinedOutput(); err != nil {
		return coreerr.E("marketplace.Installer.Update", "pull: "+strings.TrimSpace(string(output)), err)
	}

	// Reload and re-verify manifest with the same key used at install time
	medium, mErr := io.NewSandboxed(dest)
	if mErr != nil {
		return coreerr.E("marketplace.Installer.Update", "medium", mErr)
	}
	m, mErr := loadManifest(medium, installed.SignKey)
	if mErr != nil {
		return coreerr.E("marketplace.Installer.Update", "reload manifest", mErr)
	}

	// Update stored metadata
	installed.Code = safeCode
	installed.Name = m.Name
	installed.Version = m.Version
	installed.Permissions = m.Permissions

	data, err := json.Marshal(installed)
	if err != nil {
		return coreerr.E("marketplace.Installer.Update", "marshal", err)
	}

	return i.store.Set(storeGroup, safeCode, string(data))
}

// Installed returns all installed module metadata.
// Usage: Installed(...)
func (i *Installer) Installed() ([]InstalledModule, error) {
	all, err := i.store.GetAll(storeGroup)
	if err != nil {
		return nil, coreerr.E("marketplace.Installer.Installed", "list", err)
	}

	var modules []InstalledModule
	for _, raw := range all {
		var m InstalledModule
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			continue
		}
		modules = append(modules, m)
	}
	return modules, nil
}

// loadManifest loads and optionally verifies a module manifest.
func loadManifest(medium io.Medium, signKey string) (*manifest.Manifest, error) {
	if signKey != "" {
		pubBytes, err := hex.DecodeString(signKey)
		if err != nil {
			return nil, coreerr.E("marketplace.loadManifest", "decode sign key", err)
		}
		return manifest.LoadVerified(medium, ".", pubBytes)
	}
	return manifest.Load(medium, ".")
}

// gitClone clones a repository with --depth=1.
func gitClone(ctx context.Context, repo, dest string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", repo, dest)
	if output, err := cmd.CombinedOutput(); err != nil {
		return coreerr.E("marketplace.gitClone", strings.TrimSpace(string(output)), err)
	}
	return nil
}

func (i *Installer) resolveModulePath(code string) (string, string, error) {
	safeCode, dest, err := agentci.ResolvePathWithinRoot(i.modulesDir, code)
	if err != nil {
		return "", "", coreerr.E("marketplace.Installer.resolveModulePath", "resolve module path", err)
	}
	return safeCode, dest, nil
}
