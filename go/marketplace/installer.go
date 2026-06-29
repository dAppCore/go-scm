// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	"time"

	core "dappco.re/go"
	coreio "dappco.re/go/io"
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

func (i *Installer) Install(ctx context.Context, mod Module) error  /* v090-result-boundary */ {
	if i == nil {
		return core.E("marketplace.Installer.Install", "installer is required", nil)
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if i.medium == nil {
		return core.E("marketplace.Installer.Install", "medium is required", nil)
	}
	if mod.Code == "" {
		return core.E("marketplace.Installer.Install", "module code is required", nil)
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
	marshalResult := core.JSONMarshalIndent(entry, "", "  ")
	if !marshalResult.OK {
		return core.E("marketplace.Installer.Install", "encode module entry", nil)
	}
	if err := writeMediumFile(i.medium, core.PathJoin(i.modulesDir, mod.Code, sonarInstallerModuleJson), marshalResult.Value.([]byte)); err != nil {
		return err
	}
	return nil
}

func verifyModuleSignature(mod Module) error  /* v090-result-boundary */ {
	payload, err := moduleVerificationPayload(mod)
	if err != nil {
		return err
	}
	return manifest.Verify(&manifest.Manifest{
		Sign:    mod.Sign,
		SignKey: mod.SignKey,
	}, payload)
}

func moduleVerificationPayload(mod Module) ([]byte, error)  /* v090-result-boundary */ {
	cp := mod
	cp.Sign = ""
	marshalResult := core.JSONMarshal(cp)
	if !marshalResult.OK {
		return nil, core.E("marketplace.moduleVerificationPayload", "encode module", nil)
	}
	return marshalResult.Value.([]byte), nil
}

func (i *Installer) Installed() ([]InstalledModule, error)  /* v090-result-boundary */ {
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
		raw, err := readMediumFile(i.medium, core.PathJoin(i.modulesDir, entry.Name(), sonarInstallerModuleJson))
		if err != nil {
			continue
		}
		var mod InstalledModule
		if r := core.JSONUnmarshal(raw, &mod); !r.OK {
			continue
		}
		out = append(out, mod)
	}
	return out, nil
}

func (i *Installer) Remove(code string) error  /* v090-result-boundary */ {
	if i == nil || i.medium == nil {
		return core.E("marketplace.Installer.Remove", "installer is required", nil)
	}
	if err := i.medium.DeleteAll(core.PathJoin(i.modulesDir, code)); err != nil {
		return err
	}
	return nil
}

func (i *Installer) Update(ctx context.Context, code string) error  /* v090-result-boundary */ {
	if i == nil {
		return core.E("marketplace.Installer.Update", "installer is required", nil)
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if i.medium == nil {
		return core.E("marketplace.Installer.Update", "medium is required", nil)
	}
	path := core.PathJoin(i.modulesDir, code, sonarInstallerModuleJson)
	raw, err := readMediumFile(i.medium, path)
	if err != nil {
		return err
	}
	var entry InstalledModule
	if r := core.JSONUnmarshal(raw, &entry); !r.OK {
		return core.E("marketplace.Installer.Update", "decode installed module", nil)
	}
	if core.Trim(entry.Code) == "" {
		return core.E("marketplace.Installer.Update", "installed module is invalid", nil)
	}
	entry.InstalledAt = time.Now().UTC().Format(time.RFC3339Nano)
	updatedResult := core.JSONMarshalIndent(entry, "", "  ")
	if !updatedResult.OK {
		return core.E("marketplace.Installer.Update", "encode installed module", nil)
	}
	if err := writeMediumFile(i.medium, path, updatedResult.Value.([]byte)); err != nil {
		return err
	}
	return nil
}

func versionOrLatest(version string) string {
	if core.Trim(version) == "" {
		return "latest"
	}
	return core.Trim(version)
}
