// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	coreio "dappco.re/go/io"
	"dappco.re/go/scm/internal/ax/jsonx"
	"dappco.re/go/scm/manifest"
)

const (
	sonarInstallerModuleJson = "module.json"
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
}

func NewInstaller(m coreio.Medium, modulesDir string, _ ...any) *Installer {
	return &Installer{medium: m, modulesDir: modulesDir}
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
	if err := verifyModuleSignature(mod); err != nil {
		return err
	}
	entry := InstalledModule{
		Code:        mod.Code,
		Name:        mod.Name,
		Version:     versionOrLatest(mod.Version),
		Repo:        mod.Repo,
		EntryPoint:  "core.json",
		SignKey:     mod.SignKey,
		InstalledAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	raw, err := jsonx.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	if err := writeMediumFile(i.medium, filepath.Join(i.modulesDir, mod.Code, sonarInstallerModuleJson), raw); err != nil {
		return err
	}
	return nil
}

func verifyModuleSignature(mod Module) error {
	payload, err := moduleVerificationPayload(mod)
	if err != nil {
		return err
	}
	return manifest.Verify(&manifest.Manifest{
		Sign:    mod.Sign,
		SignKey: mod.SignKey,
	}, payload)
}

func moduleVerificationPayload(mod Module) ([]byte, error) {
	cp := mod
	cp.Sign = ""
	return jsonx.Marshal(cp)
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
		raw, err := readMediumFile(i.medium, filepath.Join(i.modulesDir, entry.Name(), sonarInstallerModuleJson))
		if err != nil {
			continue
		}
		var mod InstalledModule
		if err := jsonx.Unmarshal(raw, &mod); err != nil {
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
	if err := i.medium.DeleteAll(filepath.Join(i.modulesDir, code)); err != nil {
		return err
	}
	return nil
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
	if i.medium == nil {
		return errors.New("marketplace.Installer.Update: medium is required")
	}
	path := filepath.Join(i.modulesDir, code, sonarInstallerModuleJson)
	raw, err := readMediumFile(i.medium, path)
	if err != nil {
		return err
	}
	var entry InstalledModule
	if err := jsonx.Unmarshal(raw, &entry); err != nil {
		return err
	}
	if strings.TrimSpace(entry.Code) == "" {
		return errors.New("marketplace.Installer.Update: installed module is invalid")
	}
	entry.InstalledAt = time.Now().UTC().Format(time.RFC3339Nano)
	updated, err := jsonx.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	if err := writeMediumFile(i.medium, path, updated); err != nil {
		return err
	}
	return nil
}

func versionOrLatest(version string) string {
	if strings.TrimSpace(version) == "" {
		return "latest"
	}
	return strings.TrimSpace(version)
}
