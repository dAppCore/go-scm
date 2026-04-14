// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	os "dappco.re/go/core/scm/internal/ax/osx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	"sort"

	"dappco.re/go/core/cli/pkg/cli"
	coreio "dappco.re/go/core/io"
	"dappco.re/go/core/io/store"
	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/scm/agentci"
	"dappco.re/go/core/scm/marketplace"
)

func addPackageCommand(parent *cli.Command) {
	pkgCmd := &cli.Command{
		Use:   "pkg",
		Short: "Marketplace package operations",
		Long:  "Search, install, update, list, and publish packages from the git-backed marketplace.",
	}
	parent.AddCommand(pkgCmd)

	addPackageSearchCommand(pkgCmd)
	addPackageInstallCommand(pkgCmd)
	addPackageUpdateCommand(pkgCmd)
	addPackageListCommand(pkgCmd)
	addPackagePublishCommand(pkgCmd)
}

func addPackageSearchCommand(parent *cli.Command) {
	var indexPath string

	cmd := &cli.Command{
		Use:   "search [query]",
		Short: "Search the marketplace index",
		RunE: func(cmd *cli.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			return runPackageSearch(indexPath, query)
		},
	}

	cmd.Flags().StringVar(&indexPath, "index", "index.json", "Path to marketplace index.json")
	parent.AddCommand(cmd)
}

func addPackageInstallCommand(parent *cli.Command) {
	var (
		indexPath  string
		modulesDir string
		storePath  string
	)

	cmd := &cli.Command{
		Use:   "install <code[@version]>",
		Short: "Install a package from the marketplace",
		RunE: func(cmd *cli.Command, args []string) error {
			if len(args) != 1 {
				return coreerr.E("scm.runPackageInstall", "package code is required", nil)
			}
			return runPackageInstall(indexPath, modulesDir, storePath, args[0])
		},
	}

	cmd.Flags().StringVar(&indexPath, "index", "index.json", "Path to marketplace index.json")
	cmd.Flags().StringVar(&modulesDir, "modules-dir", defaultModulesDir(), "Directory for installed modules")
	cmd.Flags().StringVar(&storePath, "store", defaultModuleStorePath(), "Path to installed module store database")
	parent.AddCommand(cmd)
}

func addPackageUpdateCommand(parent *cli.Command) {
	var (
		modulesDir string
		storePath  string
	)

	cmd := &cli.Command{
		Use:   "update <code>",
		Short: "Update an installed package",
		RunE: func(cmd *cli.Command, args []string) error {
			if len(args) != 1 {
				return coreerr.E("scm.runPackageUpdate", "package code is required", nil)
			}
			return runPackageUpdate(modulesDir, storePath, args[0])
		},
	}

	cmd.Flags().StringVar(&modulesDir, "modules-dir", defaultModulesDir(), "Directory for installed modules")
	cmd.Flags().StringVar(&storePath, "store", defaultModuleStorePath(), "Path to installed module store database")
	parent.AddCommand(cmd)
}

func addPackageListCommand(parent *cli.Command) {
	var (
		modulesDir string
		storePath  string
	)

	cmd := &cli.Command{
		Use:   "list",
		Short: "List installed packages",
		RunE: func(cmd *cli.Command, args []string) error {
			return runPackageList(modulesDir, storePath)
		},
	}

	cmd.Flags().StringVar(&modulesDir, "modules-dir", defaultModulesDir(), "Directory for installed modules")
	cmd.Flags().StringVar(&storePath, "store", defaultModuleStorePath(), "Path to installed module store database")
	parent.AddCommand(cmd)
}

func addPackagePublishCommand(parent *cli.Command) {
	var (
		dirs     []string
		output   string
		forgeURL string
		org      string
	)

	cmd := &cli.Command{
		Use:   "publish",
		Short: "Publish packages into a marketplace index",
		Long:  "Scan directories for package manifests and write a marketplace index.json.",
		RunE: func(cmd *cli.Command, args []string) error {
			if len(dirs) == 0 {
				dirs = []string{"."}
			}
			return runIndex(dirs, output, forgeURL, org)
		},
	}

	cmd.Flags().StringArrayVarP(&dirs, "dir", "d", nil, "Directories to scan (repeatable, default: current directory)")
	cmd.Flags().StringVarP(&output, "output", "o", "index.json", "Output path for the index file")
	cmd.Flags().StringVar(&forgeURL, "forge-url", "", "Forge base URL for repo links")
	cmd.Flags().StringVar(&org, "org", "", "Organisation for repo links")
	parent.AddCommand(cmd)
}

func runPackageSearch(indexPath, query string) error {
	results, err := packageSearch(indexPath, query)
	if err != nil {
		return err
	}

	cli.Blank()
	if len(results) == 0 {
		cli.Print("  %s\n\n", dimStyle.Render("no packages matched"))
		return nil
	}

	for _, mod := range results {
		version := mod.Version
		if version == "" {
			version = "unversioned"
		}
		cli.Print("  %s  %s  %s\n", valueStyle.Render(mod.Code), dimStyle.Render(version), dimStyle.Render(mod.Repo))
	}
	cli.Blank()
	return nil
}

func packageSearch(indexPath, query string) ([]marketplace.Module, error) {
	idx, err := marketplace.LoadIndex(coreio.Local, indexPath)
	if err != nil {
		return nil, cli.WrapVerb(err, "load", indexPath)
	}
	if strings.TrimSpace(query) == "" {
		return idx.Modules, nil
	}
	return idx.Search(query), nil
}

func runPackageInstall(indexPath, modulesDir, storePath, request string) error {
	installed, err := packageInstall(context.Background(), indexPath, modulesDir, storePath, request)
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Print("  %s %s\n", successStyle.Render("installed"), valueStyle.Render(installed.Code))
	cli.Print("  %s %s\n", dimStyle.Render("version:"), valueStyle.Render(installed.Version))
	cli.Print("  %s %s\n", dimStyle.Render("path:"), valueStyle.Render(filepath.Join(modulesDir, installed.Code)))
	cli.Blank()
	return nil
}

func packageInstall(ctx context.Context, indexPath, modulesDir, storePath, request string) (marketplace.InstalledModule, error) {
	code, version, err := parseModuleRequest(request)
	if err != nil {
		return marketplace.InstalledModule{}, err
	}

	idx, err := marketplace.LoadIndex(coreio.Local, indexPath)
	if err != nil {
		return marketplace.InstalledModule{}, cli.WrapVerb(err, "load", indexPath)
	}

	mod, ok := idx.Find(code)
	if !ok {
		return marketplace.InstalledModule{}, coreerr.E("scm.packageInstall", "package not found: "+code, nil)
	}
	if version != "" {
		mod.Version = version
	}

	st, err := openModuleStore(storePath)
	if err != nil {
		return marketplace.InstalledModule{}, err
	}
	defer st.Close()

	inst := marketplace.NewInstaller(coreio.Local, modulesDir, st)
	if err := inst.Install(ctx, mod); err != nil {
		return marketplace.InstalledModule{}, err
	}

	return findInstalledModule(inst, code)
}

func runPackageUpdate(modulesDir, storePath, code string) error {
	installed, err := packageUpdate(context.Background(), modulesDir, storePath, code)
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Print("  %s %s\n", successStyle.Render("updated"), valueStyle.Render(installed.Code))
	cli.Print("  %s %s\n", dimStyle.Render("version:"), valueStyle.Render(installed.Version))
	cli.Blank()
	return nil
}

func packageUpdate(ctx context.Context, modulesDir, storePath, code string) (marketplace.InstalledModule, error) {
	st, err := openModuleStore(storePath)
	if err != nil {
		return marketplace.InstalledModule{}, err
	}
	defer st.Close()

	inst := marketplace.NewInstaller(coreio.Local, modulesDir, st)
	if err := inst.Update(ctx, code); err != nil {
		return marketplace.InstalledModule{}, err
	}

	return findInstalledModule(inst, code)
}

func runPackageList(modulesDir, storePath string) error {
	modules, err := packageList(modulesDir, storePath)
	if err != nil {
		return err
	}

	cli.Blank()
	if len(modules) == 0 {
		cli.Print("  %s\n\n", dimStyle.Render("no installed packages"))
		return nil
	}

	for _, mod := range modules {
		cli.Print("  %s  %s  %s\n", valueStyle.Render(mod.Code), dimStyle.Render(mod.Version), dimStyle.Render(mod.Repo))
	}
	cli.Blank()
	return nil
}

func packageList(modulesDir, storePath string) ([]marketplace.InstalledModule, error) {
	st, err := openModuleStore(storePath)
	if err != nil {
		return nil, err
	}
	defer st.Close()

	inst := marketplace.NewInstaller(coreio.Local, modulesDir, st)
	modules, err := inst.Installed()
	if err != nil {
		return nil, err
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Code < modules[j].Code
	})
	return modules, nil
}

func parseModuleRequest(request string) (string, string, error) {
	request = strings.TrimSpace(request)
	if request == "" {
		return "", "", coreerr.E("scm.parseModuleRequest", "package code is required", nil)
	}

	code := request
	version := ""
	if idx := strings.LastIndex(request, "@"); idx >= 0 {
		code = request[:idx]
		version = request[idx+1:]
		if strings.TrimSpace(version) == "" {
			return "", "", coreerr.E("scm.parseModuleRequest", "version is empty after @", nil)
		}
	}

	safeCode, err := agentci.ValidatePathElement(code)
	if err != nil {
		return "", "", coreerr.E("scm.parseModuleRequest", "invalid package code", err)
	}
	return safeCode, version, nil
}

func defaultModulesDir() string {
	return filepath.Join(defaultCoreDir(), "modules")
}

func defaultModuleStorePath() string {
	return filepath.Join(defaultCoreDir(), "scm", "modules.db")
}

func defaultCoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".core"
	}
	return filepath.Join(home, ".core")
}

func openModuleStore(storePath string) (*store.KeyValueStore, error) {
	if storePath != ":memory:" {
		if err := coreio.Local.EnsureDir(filepath.Dir(storePath)); err != nil {
			return nil, cli.WrapVerb(err, "create", filepath.Dir(storePath))
		}
	}

	st, err := store.New(store.Options{Path: storePath})
	if err != nil {
		return nil, cli.WrapVerb(err, "open", storePath)
	}
	return st, nil
}

func findInstalledModule(inst *marketplace.Installer, code string) (marketplace.InstalledModule, error) {
	modules, err := inst.Installed()
	if err != nil {
		return marketplace.InstalledModule{}, err
	}
	for _, mod := range modules {
		if mod.Code == code {
			return mod, nil
		}
	}
	return marketplace.InstalledModule{}, coreerr.E("scm.findInstalledModule", "installed package not found: "+code, nil)
}
