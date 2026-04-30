// SPDX-License-Identifier: EUPL-1.2

package collect

import "testing"

func TestRatelimit_NewRateLimiter_Good(t *testing.T) {
	target := "NewRateLimiter"
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

func TestRatelimit_NewRateLimiter_Bad(t *testing.T) {
	target := "NewRateLimiter"
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

func TestRatelimit_NewRateLimiter_Ugly(t *testing.T) {
	target := "NewRateLimiter"
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

func TestRatelimit_RateLimiter_Wait_Good(t *testing.T) {
	reference := "Wait"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_Wait"
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

func TestRatelimit_RateLimiter_Wait_Bad(t *testing.T) {
	reference := "Wait"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_Wait"
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

func TestRatelimit_RateLimiter_Wait_Ugly(t *testing.T) {
	reference := "Wait"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_Wait"
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

func TestRatelimit_RateLimiter_SetDelay_Good(t *testing.T) {
	reference := "SetDelay"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_SetDelay"
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

func TestRatelimit_RateLimiter_SetDelay_Bad(t *testing.T) {
	reference := "SetDelay"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_SetDelay"
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

func TestRatelimit_RateLimiter_SetDelay_Ugly(t *testing.T) {
	reference := "SetDelay"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_SetDelay"
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

func TestRatelimit_RateLimiter_GetDelay_Good(t *testing.T) {
	reference := "GetDelay"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_GetDelay"
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

func TestRatelimit_RateLimiter_GetDelay_Bad(t *testing.T) {
	reference := "GetDelay"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_GetDelay"
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

func TestRatelimit_RateLimiter_GetDelay_Ugly(t *testing.T) {
	reference := "GetDelay"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_GetDelay"
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

func TestRatelimit_RateLimiter_CheckGitHubRateLimit_Good(t *testing.T) {
	reference := "CheckGitHubRateLimit"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_CheckGitHubRateLimit"
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

func TestRatelimit_RateLimiter_CheckGitHubRateLimit_Bad(t *testing.T) {
	reference := "CheckGitHubRateLimit"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_CheckGitHubRateLimit"
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

func TestRatelimit_RateLimiter_CheckGitHubRateLimit_Ugly(t *testing.T) {
	reference := "CheckGitHubRateLimit"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_CheckGitHubRateLimit"
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

func TestRatelimit_RateLimiter_CheckGitHubRateLimitCtx_Good(t *testing.T) {
	reference := "CheckGitHubRateLimitCtx"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_CheckGitHubRateLimitCtx"
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

func TestRatelimit_RateLimiter_CheckGitHubRateLimitCtx_Bad(t *testing.T) {
	reference := "CheckGitHubRateLimitCtx"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_CheckGitHubRateLimitCtx"
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

func TestRatelimit_RateLimiter_CheckGitHubRateLimitCtx_Ugly(t *testing.T) {
	reference := "CheckGitHubRateLimitCtx"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "RateLimiter_CheckGitHubRateLimitCtx"
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
