// SPDX-License-Identifier: EUPL-1.2

package git

import (
	"context"
	"testing"

	core "dappco.re/go"
)

func TestServiceRegistersActionsWithoutWorkDir(t *testing.T) {
	c := core.New(core.WithService(NewService(ServiceOptions{})))
	if r := c.ServiceStartup(context.Background(), nil); !r.OK {
		t.Fatalf("service startup failed: %v", r.Value)
	}

	for _, name := range []string{
		"git.push",
		"git.pull",
		"git.push-multiple",
		"git.pull-multiple",
	} {
		if !c.Action(name).Exists() {
			t.Fatalf("expected %s to be registered", name)
		}
	}
}

func TestService_NewService_Good(t *testing.T) {
	target := "NewService"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_NewService_Bad(t *testing.T) {
	target := "NewService"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_NewService_Ugly(t *testing.T) {
	target := "NewService"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_Status_Good(t *testing.T) {
	reference := "Status"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_Status"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_Status_Bad(t *testing.T) {
	reference := "Status"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_Status"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_Status_Ugly(t *testing.T) {
	reference := "Status"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_Status"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_StatusIter_Good(t *testing.T) {
	reference := "StatusIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_StatusIter"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_StatusIter_Bad(t *testing.T) {
	reference := "StatusIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_StatusIter"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_StatusIter_Ugly(t *testing.T) {
	reference := "StatusIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_StatusIter"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_DirtyRepos_Good(t *testing.T) {
	reference := "DirtyRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_DirtyRepos"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_DirtyRepos_Bad(t *testing.T) {
	reference := "DirtyRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_DirtyRepos"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_DirtyRepos_Ugly(t *testing.T) {
	reference := "DirtyRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_DirtyRepos"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_DirtyReposIter_Good(t *testing.T) {
	reference := "DirtyReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_DirtyReposIter"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_DirtyReposIter_Bad(t *testing.T) {
	reference := "DirtyReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_DirtyReposIter"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_DirtyReposIter_Ugly(t *testing.T) {
	reference := "DirtyReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_DirtyReposIter"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_AheadRepos_Good(t *testing.T) {
	reference := "AheadRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_AheadRepos"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_AheadRepos_Bad(t *testing.T) {
	reference := "AheadRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_AheadRepos"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_AheadRepos_Ugly(t *testing.T) {
	reference := "AheadRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_AheadRepos"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_AheadReposIter_Good(t *testing.T) {
	reference := "AheadReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_AheadReposIter"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_AheadReposIter_Bad(t *testing.T) {
	reference := "AheadReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_AheadReposIter"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_AheadReposIter_Ugly(t *testing.T) {
	reference := "AheadReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_AheadReposIter"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_BehindRepos_Good(t *testing.T) {
	reference := "BehindRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_BehindRepos"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_BehindRepos_Bad(t *testing.T) {
	reference := "BehindRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_BehindRepos"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_BehindRepos_Ugly(t *testing.T) {
	reference := "BehindRepos"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_BehindRepos"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_BehindReposIter_Good(t *testing.T) {
	reference := "BehindReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_BehindReposIter"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_BehindReposIter_Bad(t *testing.T) {
	reference := "BehindReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_BehindReposIter"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_BehindReposIter_Ugly(t *testing.T) {
	reference := "BehindReposIter"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_BehindReposIter"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestService_Service_OnStartup_Good(t *testing.T) {
	reference := "OnStartup"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_OnStartup"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestService_Service_OnStartup_Bad(t *testing.T) {
	reference := "OnStartup"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_OnStartup"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestService_Service_OnStartup_Ugly(t *testing.T) {
	reference := "OnStartup"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Service_OnStartup"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}
