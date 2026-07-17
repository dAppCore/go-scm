// SPDX-License-Identifier: EUPL-1.2

package forge

import (
	core "dappco.re/go"
	forgejo "dappco.re/go/scm/third_party/forgejo/forgejo"
)

func TestForge_hasNextForgePage_Good(t *core.T) {
	core.AssertTrue(t, hasNextForgePage(&forgeResponse{LastPage: 3}, 1))
}

func TestForge_hasNextForgePage_Bad(t *core.T) {
	core.AssertFalse(t, hasNextForgePage(nil, 1))
}

func TestForge_hasNextForgePage_Ugly(t *core.T) {
	core.AssertFalse(t, hasNextForgePage(&forgeResponse{LastPage: 2}, 2))
}

func TestForge_hasMoreForgeItems_Good(t *core.T) {
	items := []int{1, 2, 3}
	core.AssertTrue(t, hasMoreForgeItems(items, &forgeResponse{LastPage: 5}, 1, 3))
}

func TestForge_hasMoreForgeItems_Bad(t *core.T) {
	// A short page (fewer items than the limit) means there is no more.
	items := []int{1}
	core.AssertFalse(t, hasMoreForgeItems(items, &forgeResponse{LastPage: 5}, 1, 3))
}

func TestForge_hasMoreForgeItems_Ugly(t *core.T) {
	// A full page with no response metadata is treated as "maybe more".
	items := []int{1, 2, 3}
	core.AssertTrue(t, hasMoreForgeItems(items, nil, 1, 3))
}

func TestForge_appendUniqueLabels_Good(t *core.T) {
	seen := map[string]struct{}{}
	got := appendUniqueLabels(nil, seen, []*forgejo.Label{
		{Name: "bug"},
		{Name: "feature"},
	})
	core.AssertLen(t, got, 2)
}

func TestForge_appendUniqueLabels_Bad(t *core.T) {
	seen := map[string]struct{}{}
	got := appendUniqueLabels(nil, seen, nil)
	core.AssertEmpty(t, got)
}

func TestForge_appendUniqueLabels_Ugly(t *core.T) {
	// Case-insensitive dedup: "Bug" and "bug" collapse to one entry.
	seen := map[string]struct{}{}
	got := appendUniqueLabels(nil, seen, []*forgejo.Label{
		{Name: "Bug"},
		{Name: "bug"},
	})
	core.AssertLen(t, got, 1)
}
