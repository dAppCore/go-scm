// SPDX-License-Identifier: EUPL-1.2

package marketplace

import (
	"context"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"

	"dappco.re/go/core/scm/git"
)

func moduleVersionForRepo(repoPath, fallback string) string {
	tag := latestRepoSemverTag(repoPath)
	if tag == "" {
		return fallback
	}
	return strings.TrimPrefix(tag, "v")
}

func latestRepoSemverTag(repoPath string) string {
	if strings.TrimSpace(repoPath) == "" {
		return ""
	}

	tags, err := git.ListRemoteTags(context.Background(), repoPath)
	if err != nil {
		return ""
	}
	return latestSemverTag(tags)
}
