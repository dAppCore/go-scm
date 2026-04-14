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
	"dappco.re/go/core/scm/git"
	"dappco.re/go/core/scm/manifest"
	"golang.org/x/mod/semver"
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
	ref, err := resolveModuleRef(ctx, mod)
	if err != nil {
		return coreerr.E("marketplace.Installer.Install", "resolve version", err)
	}
	if err := git.Clone(ctx, mod.Repo, dest, ref); err != nil {
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
	resolvedSignKey := strings.TrimSpace(mod.SignKey)
	if resolvedSignKey == "" {
		resolvedSignKey = strings.TrimSpace(m.SignKey)
	}
	installed := InstalledModule{
		Code:        safeCode,
		Name:        m.Name,
		Version:     installedModuleVersion(ref, m.Version),
		Repo:        mod.Repo,
		EntryPoint:  entryPoint,
		Permissions: m.Permissions,
		SignKey:     resolvedSignKey,
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
	latestRef, refErr := resolveModuleRef(ctx, Module{Repo: installed.Repo})
	if refErr == nil && latestRef != "" {
		currentRef, tagErr := git.CurrentTag(ctx, dest)
		if tagErr != nil {
			return coreerr.E("marketplace.Installer.Update", "current tag", tagErr)
		}

		if currentRef != latestRef {
			if err := git.FetchTags(ctx, dest); err != nil {
				return coreerr.E("marketplace.Installer.Update", "fetch tags", err)
			}
			if err := git.Checkout(ctx, dest, latestRef); err != nil {
				return coreerr.E("marketplace.Installer.Update", "checkout "+latestRef, err)
			}
		}
	} else {
		if output, err := cmd.CombinedOutput(); err != nil {
			return coreerr.E("marketplace.Installer.Update", "pull: "+strings.TrimSpace(string(output)), err)
		}
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
	currentTag, tagErr := git.CurrentTag(ctx, dest)
	if tagErr != nil {
		return coreerr.E("marketplace.Installer.Update", "current tag", tagErr)
	}

	// Update stored metadata
	installed.Code = safeCode
	installed.Name = m.Name
	installed.Version = installedModuleVersion(currentTag, m.Version)
	installed.Permissions = m.Permissions
	if installed.SignKey == "" {
		installed.SignKey = strings.TrimSpace(m.SignKey)
	}

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

	m, err := manifest.Load(medium, ".")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(m.SignKey) == "" {
		return m, nil
	}

	ok, err := manifest.Verify(m, nil)
	if err != nil {
		return nil, coreerr.E("marketplace.loadManifest", "verify embedded sign key", err)
	}
	if !ok {
		return nil, coreerr.E("marketplace.loadManifest", "signature verification failed for "+m.Code, nil)
	}
	return m, nil
}

func (i *Installer) resolveModulePath(code string) (string, string, error) {
	safeCode, dest, err := agentci.ResolvePathWithinRoot(i.modulesDir, code)
	if err != nil {
		return "", "", coreerr.E("marketplace.Installer.resolveModulePath", "resolve module path", err)
	}
	return safeCode, dest, nil
}

func resolveModuleRef(ctx context.Context, mod Module) (string, error) {
	tags, err := git.ListRemoteTags(ctx, mod.Repo)
	if err != nil {
		if mod.Version != "" {
			return "", nil
		}
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}

	if mod.Version != "" {
		return matchRequestedTag(tags, mod.Version)
	}

	return latestSemverTag(tags), nil
}

func matchRequestedTag(tags []string, requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	if requested == "" {
		return "", nil
	}

	candidates := []string{requested}
	if !strings.HasPrefix(requested, "v") {
		candidates = append([]string{"v" + requested}, candidates...)
	}

	for _, candidate := range candidates {
		for _, tag := range tags {
			if tag == candidate {
				return tag, nil
			}
		}
	}

	return "", coreerr.E("marketplace.matchRequestedTag", "tag not found: "+requested, nil)
}

func latestSemverTag(tags []string) string {
	type semverTag struct {
		raw       string
		canonical string
	}

	var versions []semverTag
	for _, tag := range tags {
		canonical := tag
		if !semver.IsValid(canonical) && semver.IsValid("v"+canonical) {
			canonical = "v" + canonical
		}
		if !semver.IsValid(canonical) {
			continue
		}
		versions = append(versions, semverTag{
			raw:       tag,
			canonical: canonical,
		})
	}

	if len(versions) == 0 {
		return ""
	}

	best := versions[0]
	for _, version := range versions[1:] {
		if semver.Compare(version.canonical, best.canonical) > 0 {
			best = version
		}
	}
	return best.raw
}

func installedModuleVersion(ref, fallback string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return fallback
	}

	canonical := ref
	if !semver.IsValid(canonical) && semver.IsValid("v"+canonical) {
		canonical = "v" + canonical
	}
	if semver.IsValid(canonical) {
		return strings.TrimPrefix(canonical, "v")
	}

	return fallback
}
