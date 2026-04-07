// SPDX-License-Identifier: EUPL-1.2

package syncutil

import (
	"net/url"

	coreerr "dappco.re/go/core/log"
	"dappco.re/go/core/scm/agentci"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
)

// ParseRepoName normalises a sync argument into a validated repo name.
// Usage: ParseRepoName(...)
func ParseRepoName(arg string) (string, error) {
	decoded, err := url.PathUnescape(arg)
	if err != nil {
		return "", coreerr.E("syncutil.ParseRepoName", "decode repo argument", err)
	}

	parts := strings.Split(decoded, "/")
	switch len(parts) {
	case 1:
		name, err := agentci.ValidatePathElement(parts[0])
		if err != nil {
			return "", coreerr.E("syncutil.ParseRepoName", "invalid repo name", err)
		}
		return name, nil
	case 2:
		if _, err := agentci.ValidatePathElement(parts[0]); err != nil {
			return "", coreerr.E("syncutil.ParseRepoName", "invalid repo owner", err)
		}
		name, err := agentci.ValidatePathElement(parts[1])
		if err != nil {
			return "", coreerr.E("syncutil.ParseRepoName", "invalid repo name", err)
		}
		return name, nil
	default:
		return "", coreerr.E("syncutil.ParseRepoName", "repo argument must be repo or owner/repo", nil)
	}
}
