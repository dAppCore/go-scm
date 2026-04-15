// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	coreio "dappco.re/go/core/io"
	"dappco.re/go/scm/marketplace"
	"dappco.re/go/scm/repos"
	"github.com/gin-gonic/gin"
)

func TestScmProviderRoutesExposeState(t *testing.T) {
	gin.SetMode(gin.TestMode)

	medium := coreio.NewMockMedium()
	installer := marketplace.NewInstaller(medium, "modules", nil)
	if err := installer.Install(nil, marketplace.Module{
		Code:    "go-io",
		Name:    "Core I/O",
		Repo:    "ssh://git@example.com/core/go-io.git",
		SignKey: "ed25519:public-key",
	}); err != nil {
		t.Fatalf("install module: %v", err)
	}

	provider := NewProvider(
		&marketplace.Index{
			Version: 1,
			Modules: []marketplace.Module{{
				Code:    "go-io",
				Name:    "Core I/O",
				Repo:    "ssh://git@example.com/core/go-io.git",
				SignKey: "ed25519:public-key",
			}},
		},
		installer,
		&repos.Registry{
			Version:  1,
			BasePath: "/workspace",
			Repos: map[string]*repos.Repo{
				"go-io": {Path: "/workspace/core/go-io"},
			},
		},
		nil,
	)

	router := gin.New()
	group := router.Group(provider.BasePath())
	provider.RegisterRoutes(group)

	assertRouteOK(t, router, "/scm/health", "ok")
	assertRouteOK(t, router, "/scm/marketplace", "go-io")
	assertRouteOK(t, router, "/scm/marketplace/go-io", "Core I/O")
	assertRouteOK(t, router, "/scm/repos", "go-io")
	assertRouteOK(t, router, "/scm/repos/go-io", "/workspace/core/go-io")
	assertRouteOK(t, router, "/scm/modules", "go-io")
	assertRouteOK(t, router, "/scm/modules/go-io", "Core I/O")
}

func TestScmProviderDescribeIncludesReadOnlyRoutes(t *testing.T) {
	provider := NewProvider(nil, nil, nil, nil)
	descs := provider.Describe()

	want := map[string]bool{
		"GET /health":         false,
		"GET /marketplace":    false,
		"GET /repos":          false,
		"GET /modules":        false,
	}
	for _, desc := range descs {
		key := desc.Method + " " + desc.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}
	for key, seen := range want {
		if !seen {
			t.Fatalf("expected route description for %s", key)
		}
	}
}

func assertRouteOK(t *testing.T, router *gin.Engine, path, want string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s: expected 200, got %d with body %s", path, rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), want) {
		t.Fatalf("%s: expected body to contain %q, got %s", path, want, rec.Body.String())
	}
}
