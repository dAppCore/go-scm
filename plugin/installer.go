package plugin

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/scm/agentci"
)

// Installer handles plugin installation from GitHub.
type Installer struct {
	medium   io.Medium
	registry *Registry
}

// NewInstaller creates a new plugin installer.
func NewInstaller(m io.Medium, registry *Registry) *Installer {
	return &Installer{
		medium:   m,
		registry: registry,
	}
}

// Install downloads and installs a plugin from GitHub.
// The source format is "org/repo" or "org/repo@version".
func (i *Installer) Install(ctx context.Context, source string) error {
	org, repo, version, err := ParseSource(source)
	if err != nil {
		return coreerr.E("plugin.Installer.Install", "invalid source", err)
	}

	// Check if already installed
	if _, exists := i.registry.Get(repo); exists {
		return coreerr.E("plugin.Installer.Install", "plugin already installed: "+repo, nil)
	}

	// Clone the repository
	_, pluginDir, err := i.resolvePluginPath(repo)
	if err != nil {
		return coreerr.E("plugin.Installer.Install", "invalid plugin path", err)
	}
	if err := i.medium.EnsureDir(pluginDir); err != nil {
		return coreerr.E("plugin.Installer.Install", "failed to create plugin directory", err)
	}

	if err := i.cloneRepo(ctx, org, repo, version, pluginDir); err != nil {
		return coreerr.E("plugin.Installer.Install", "failed to clone repository", err)
	}

	// Load and validate manifest
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	manifest, err := LoadManifest(i.medium, manifestPath)
	if err != nil {
		// Clean up on failure
		_ = i.medium.DeleteAll(pluginDir)
		return coreerr.E("plugin.Installer.Install", "failed to load manifest", err)
	}

	if err := manifest.Validate(); err != nil {
		_ = i.medium.DeleteAll(pluginDir)
		return coreerr.E("plugin.Installer.Install", "invalid manifest", err)
	}

	// Resolve version
	if version == "" {
		version = manifest.Version
	}

	// Register in the registry
	cfg := &PluginConfig{
		Name:        manifest.Name,
		Version:     version,
		Source:      fmt.Sprintf("github:%s/%s", org, repo),
		Enabled:     true,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}

	if err := i.registry.Add(cfg); err != nil {
		return coreerr.E("plugin.Installer.Install", "failed to register plugin", err)
	}

	if err := i.registry.Save(); err != nil {
		return coreerr.E("plugin.Installer.Install", "failed to save registry", err)
	}

	return nil
}

// Update updates a plugin to the latest version.
func (i *Installer) Update(ctx context.Context, name string) error {
	safeName, pluginDir, err := i.resolvePluginPath(name)
	if err != nil {
		return coreerr.E("plugin.Installer.Update", "invalid plugin name", err)
	}

	cfg, ok := i.registry.Get(safeName)
	if !ok {
		return coreerr.E("plugin.Installer.Update", "plugin not found: "+safeName, nil)
	}

	// Pull latest changes
	cmd := exec.CommandContext(ctx, "git", "-C", pluginDir, "pull", "--ff-only")
	if output, err := cmd.CombinedOutput(); err != nil {
		return coreerr.E("plugin.Installer.Update", "failed to pull updates: "+strings.TrimSpace(string(output)), err)
	}

	// Reload manifest to get updated version
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	manifest, err := LoadManifest(i.medium, manifestPath)
	if err != nil {
		return coreerr.E("plugin.Installer.Update", "failed to read updated manifest", err)
	}

	// Update registry
	cfg.Version = manifest.Version
	if err := i.registry.Save(); err != nil {
		return coreerr.E("plugin.Installer.Update", "failed to save registry", err)
	}

	return nil
}

// Remove uninstalls a plugin by removing its files and registry entry.
func (i *Installer) Remove(name string) error {
	safeName, pluginDir, err := i.resolvePluginPath(name)
	if err != nil {
		return coreerr.E("plugin.Installer.Remove", "invalid plugin name", err)
	}

	if _, ok := i.registry.Get(safeName); !ok {
		return coreerr.E("plugin.Installer.Remove", "plugin not found: "+safeName, nil)
	}

	// Delete plugin directory
	if i.medium.Exists(pluginDir) {
		if err := i.medium.DeleteAll(pluginDir); err != nil {
			return coreerr.E("plugin.Installer.Remove", "failed to delete plugin files", err)
		}
	}

	// Remove from registry
	if err := i.registry.Remove(safeName); err != nil {
		return coreerr.E("plugin.Installer.Remove", "failed to unregister plugin", err)
	}

	if err := i.registry.Save(); err != nil {
		return coreerr.E("plugin.Installer.Remove", "failed to save registry", err)
	}

	return nil
}

// cloneRepo clones a GitHub repository using the gh CLI.
func (i *Installer) cloneRepo(ctx context.Context, org, repo, version, dest string) error {
	repoURL := fmt.Sprintf("%s/%s", org, repo)

	args := []string{"repo", "clone", repoURL, dest}
	if version != "" {
		args = append(args, "--", "--branch", version)
	}

	cmd := exec.CommandContext(ctx, "gh", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return coreerr.E("plugin.Installer.cloneRepo", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// ParseSource parses a plugin source string into org, repo, and version.
// Accepted formats:
//   - "org/repo" -> org="org", repo="repo", version=""
//   - "org/repo@v1.0" -> org="org", repo="repo", version="v1.0"
func ParseSource(source string) (org, repo, version string, err error) {
	source, err = url.PathUnescape(source)
	if err != nil {
		return "", "", "", coreerr.E("plugin.ParseSource", "invalid source path", err)
	}
	if source == "" {
		return "", "", "", coreerr.E("plugin.ParseSource", "source is empty", nil)
	}

	// Split off version if present
	atIdx := strings.LastIndex(source, "@")
	path := source
	if atIdx != -1 {
		path = source[:atIdx]
		version = source[atIdx+1:]
		if version == "" {
			return "", "", "", coreerr.E("plugin.ParseSource", "version is empty after @", nil)
		}
	}

	// Split org/repo
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", "", coreerr.E("plugin.ParseSource", "source must be in format org/repo[@version]", nil)
	}

	org, err = agentci.ValidatePathElement(parts[0])
	if err != nil {
		return "", "", "", coreerr.E("plugin.ParseSource", "invalid org", err)
	}
	repo, err = agentci.ValidatePathElement(parts[1])
	if err != nil {
		return "", "", "", coreerr.E("plugin.ParseSource", "invalid repo", err)
	}

	return org, repo, version, nil
}

func (i *Installer) resolvePluginPath(name string) (string, string, error) {
	safeName, path, err := agentci.ResolvePathWithinRoot(i.registry.basePath, name)
	if err != nil {
		return "", "", coreerr.E("plugin.Installer.resolvePluginPath", "resolve plugin path", err)
	}
	return safeName, path, nil
}
