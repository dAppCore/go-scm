// SPDX-License-Identifier: EUPL-1.2

package scm

import (
	"context"
	"strconv"

	filepath "dappco.re/go/core/scm/internal/ax/filepathx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"

	"dappco.re/go/core/cli/pkg/cli"
	coreio "dappco.re/go/core/io"
	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/scm/git"
	"dappco.re/go/core/scm/repos"
)

const defaultWorkspaceCommitMessage = "chore: sync workspace"

func addDevCommand(parent *cli.Command) {
	devCmd := &cli.Command{
		Use:   "dev",
		Short: "Workspace development operations",
		Long:  "Inspect and operate on repos.yaml workspaces across discovered SCM registries.",
	}
	parent.AddCommand(devCmd)

	addDevHealthCommand(devCmd)
	addDevPullCommand(devCmd)
	addDevPushCommand(devCmd)
	addDevCommitCommand(devCmd)
	addDevWorkCommand(devCmd)
	addDevImpactCommand(devCmd)
}

func addDevHealthCommand(parent *cli.Command) {
	var registries []string

	cmd := &cli.Command{
		Use:   "health",
		Short: "Show workspace health across repositories",
		RunE: func(cmd *cli.Command, args []string) error {
			return runDevHealth(registries)
		},
	}

	cmd.Flags().StringArrayVar(&registries, "registry", nil, "Explicit repos.yaml paths (repeatable)")
	parent.AddCommand(cmd)
}

func addDevPullCommand(parent *cli.Command) {
	var registries []string

	cmd := &cli.Command{
		Use:   "pull",
		Short: "Pull repositories that are behind upstream",
		RunE: func(cmd *cli.Command, args []string) error {
			return runDevPull(registries)
		},
	}

	cmd.Flags().StringArrayVar(&registries, "registry", nil, "Explicit repos.yaml paths (repeatable)")
	parent.AddCommand(cmd)
}

func addDevPushCommand(parent *cli.Command) {
	var registries []string

	cmd := &cli.Command{
		Use:   "push",
		Short: "Push repositories with unpushed commits",
		RunE: func(cmd *cli.Command, args []string) error {
			return runDevPush(registries)
		},
	}

	cmd.Flags().StringArrayVar(&registries, "registry", nil, "Explicit repos.yaml paths (repeatable)")
	parent.AddCommand(cmd)
}

func addDevCommitCommand(parent *cli.Command) {
	var (
		registries []string
		message    string
	)

	cmd := &cli.Command{
		Use:   "commit",
		Short: "Commit dirty repositories with a shared message",
		RunE: func(cmd *cli.Command, args []string) error {
			return runDevCommit(registries, message)
		},
	}

	cmd.Flags().StringArrayVar(&registries, "registry", nil, "Explicit repos.yaml paths (repeatable)")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message to use for dirty repositories (defaults to a workspace sync message)")
	parent.AddCommand(cmd)
}

func addDevWorkCommand(parent *cli.Command) {
	var (
		registries []string
		message    string
	)

	cmd := &cli.Command{
		Use:   "work",
		Short: "Run the workspace status, commit, and push workflow",
		RunE: func(cmd *cli.Command, args []string) error {
			return runDevWork(registries, message)
		},
	}

	cmd.Flags().StringArrayVar(&registries, "registry", nil, "Explicit repos.yaml paths (repeatable)")
	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message for dirty repositories (defaults to a workspace sync message)")
	parent.AddCommand(cmd)
}

func addDevImpactCommand(parent *cli.Command) {
	var registries []string

	cmd := &cli.Command{
		Use:   "impact <repo>",
		Short: "Show transitive dependency impact for a repository",
		RunE: func(cmd *cli.Command, args []string) error {
			if len(args) != 1 {
				return coreerr.E("scm.runDevImpact", "repo name is required", nil)
			}
			return runDevImpact(registries, args[0])
		},
	}

	cmd.Flags().StringArrayVar(&registries, "registry", nil, "Explicit repos.yaml paths (repeatable)")
	parent.AddCommand(cmd)
}

func runDevHealth(registryPaths []string) error {
	statuses, err := workspaceStatuses(context.Background(), registryPaths)
	if err != nil {
		return err
	}

	cli.Blank()
	for _, status := range statuses {
		state := "clean"
		if status.IsDirty() {
			state = "dirty"
		}
		cli.Print(
			"  %s  %s  %s  ahead=%s  behind=%s\n",
			valueStyle.Render(status.Name),
			dimStyle.Render(status.Branch),
			dimStyle.Render(state),
			numberStyle.Render(strconv.Itoa(status.Ahead)),
			numberStyle.Render(strconv.Itoa(status.Behind)),
		)
	}
	cli.Blank()
	return nil
}

func runDevPull(registryPaths []string) error {
	statuses, err := workspaceStatuses(context.Background(), registryPaths)
	if err != nil {
		return err
	}

	var failures []string
	var pulled int
	for _, status := range statuses {
		if status.Error != nil || !status.HasUnpulled() {
			continue
		}
		if err := git.Pull(context.Background(), status.Path); err != nil {
			failures = append(failures, status.Name+": "+err.Error())
			continue
		}
		pulled++
	}

	return printDevSummary("pulled", pulled, failures)
}

func runDevPush(registryPaths []string) error {
	statuses, err := workspaceStatuses(context.Background(), registryPaths)
	if err != nil {
		return err
	}

	var paths []string
	names := make(map[string]string)
	for _, status := range statuses {
		if status.Error != nil || !status.HasUnpushed() {
			continue
		}
		paths = append(paths, status.Path)
		names[status.Path] = status.Name
	}

	results := git.PushMultiple(context.Background(), paths, names)
	var failures []string
	var pushed int
	for _, result := range results {
		if result.Success {
			pushed++
			continue
		}
		if result.Error != nil {
			failures = append(failures, result.Name+": "+result.Error.Error())
		}
	}

	return printDevSummary("pushed", pushed, failures)
}

func runDevCommit(registryPaths []string, message string) error {
	commitMessage := strings.TrimSpace(message)
	if commitMessage == "" {
		commitMessage = defaultWorkspaceCommitMessage
	}

	statuses, err := workspaceStatuses(context.Background(), registryPaths)
	if err != nil {
		return err
	}

	var failures []string
	var committed int
	for _, status := range statuses {
		if status.Error != nil || !status.IsDirty() {
			continue
		}
		if err := git.AddAll(context.Background(), status.Path); err != nil {
			failures = append(failures, status.Name+": "+err.Error())
			continue
		}
		if err := git.Commit(context.Background(), status.Path, commitMessage); err != nil {
			failures = append(failures, status.Name+": "+err.Error())
			continue
		}
		committed++
	}

	return printDevSummary("committed", committed, failures)
}

func runDevWork(registryPaths []string, message string) error {
	if err := runDevHealth(registryPaths); err != nil {
		return err
	}
	if err := runDevCommit(registryPaths, message); err != nil {
		return err
	}
	return runDevPush(registryPaths)
}

func runDevImpact(registryPaths []string, target string) error {
	impacted, err := workspaceImpact(registryPaths, target)
	if err != nil {
		return err
	}

	cli.Blank()
	if len(impacted) == 0 {
		cli.Print("  %s\n\n", dimStyle.Render("no dependent repositories"))
		return nil
	}
	for _, repo := range impacted {
		cli.Print("  %s  %s\n", valueStyle.Render(repo.Name), dimStyle.Render(repo.Path))
	}
	cli.Blank()
	return nil
}

func workspaceStatuses(ctx context.Context, registryPaths []string) ([]git.RepoStatus, error) {
	repoList, err := loadWorkspaceRepos(registryPaths)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(repoList))
	names := make(map[string]string, len(repoList))
	for _, repo := range repoList {
		paths = append(paths, repo.Path)
		names[repo.Path] = repo.Name
	}

	return git.Status(ctx, git.StatusOptions{
		Paths: paths,
		Names: names,
	}), nil
}

func workspaceImpact(registryPaths []string, target string) ([]*repos.Repo, error) {
	regs, err := loadWorkspaceRegistries(registryPaths)
	if err != nil {
		return nil, err
	}

	merged := repos.MergeRegistries(regs...)
	return merged.Impact(target)
}

func loadWorkspaceRepos(registryPaths []string) ([]*repos.Repo, error) {
	regs, err := loadWorkspaceRegistries(registryPaths)
	if err != nil {
		return nil, err
	}

	merged := repos.MergeRegistries(regs...)
	return merged.List(), nil
}

func loadWorkspaceRegistries(registryPaths []string) ([]*repos.Registry, error) {
	if len(registryPaths) == 0 {
		return repos.LoadRegistries(coreio.Local)
	}

	regs := make([]*repos.Registry, 0, len(registryPaths))
	for _, path := range registryPaths {
		reg, err := repos.LoadRegistry(coreio.Local, path)
		if err != nil {
			return nil, cli.WrapVerb(err, "load", filepath.Base(path))
		}
		regs = append(regs, reg)
	}
	return regs, nil
}

func printDevSummary(action string, count int, failures []string) error {
	cli.Blank()
	cli.Print("  %s %s\n", successStyle.Render(action), numberStyle.Render(strconv.Itoa(count)))
	for _, failure := range failures {
		cli.Print("  %s %s\n", errorStyle.Render("error"), dimStyle.Render(failure))
	}
	cli.Blank()

	if len(failures) > 0 {
		return coreerr.E("scm.printDevSummary", strings.Join(failures, "; "), nil)
	}
	return nil
}
