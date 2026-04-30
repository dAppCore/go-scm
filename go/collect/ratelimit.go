// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	// Note: context.Context is retained as the rate limiter cancellation contract.
	"context"
	// Note: AX-6 — structural boundary: gh CLI rate-limit probing intentionally invokes a binary until collect has a process-service boundary.
	`os/exec`
	// Note: strconv.Atoi is retained for parsing gh rate-limit output.
	"strconv"
	// Note: sync.Mutex protects limiter state and has no core equivalent.
	"sync"
	// Note: time is retained for limiter delay calculations and timers.
	"time"

	core "dappco.re/go"
)

const (
	sonarRatelimitCollectRatelimiterCheckgithubratelimitctx = "collect.RateLimiter.CheckGitHubRateLimitCtx"
)

// RateLimiter tracks per-source rate limiting to avoid overwhelming APIs.
type RateLimiter struct {
	mu     sync.Mutex
	delays map[string]time.Duration
	last   map[string]time.Time
}

var defaultDelays = map[string]time.Duration{
	"github":      500 * time.Millisecond,
	"bitcointalk": 2 * time.Second,
	"coingecko":   1500 * time.Millisecond,
	"iacr":        1 * time.Second,
	"arxiv":       1 * time.Second,
}

// NewRateLimiter creates a limiter with default delays.
func NewRateLimiter() *RateLimiter {
	delays := make(map[string]time.Duration, len(defaultDelays))
	for k, v := range defaultDelays {
		delays[k] = v
	}
	return &RateLimiter{
		delays: delays,
		last:   make(map[string]time.Time),
	}
}

// Wait blocks until the rate limit allows the next request for the given source.
// It respects context cancellation.
func (r *RateLimiter) Wait(ctx context.Context, source string) error  /* v090-result-boundary */ {
	if r == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	r.mu.Lock()
	delay, ok := r.delays[source]
	if !ok {
		delay = 500 * time.Millisecond
	}
	lastTime := r.last[source]
	elapsed := time.Since(lastTime)
	if elapsed >= delay {
		r.last[source] = time.Now()
		r.mu.Unlock()
		return nil
	}
	remaining := delay - elapsed
	r.mu.Unlock()

	timer := time.NewTimer(remaining)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}

	r.mu.Lock()
	r.last[source] = time.Now()
	r.mu.Unlock()
	return nil
}

// SetDelay sets the delay for a source.
func (r *RateLimiter) SetDelay(source string, d time.Duration) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.delays == nil {
		r.delays = make(map[string]time.Duration)
	}
	r.delays[source] = d
}

// GetDelay returns the delay configured for a source.
func (r *RateLimiter) GetDelay(source string) time.Duration {
	if r == nil {
		return 500 * time.Millisecond
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if d, ok := r.delays[source]; ok {
		return d
	}
	return 500 * time.Millisecond
}

// CheckGitHubRateLimit checks GitHub API rate limit status via gh api.
// Returns used and limit counts. Auto-pauses at 75% usage by increasing
// the GitHub rate limit delay.
//
// Deprecated: Use CheckGitHubRateLimitCtx for context-aware cancellation.
func (r *RateLimiter) CheckGitHubRateLimit() (used, limit int, err error)  /* v090-result-boundary */ {
	return r.CheckGitHubRateLimitCtx(context.Background())
}

// CheckGitHubRateLimitCtx checks GitHub API rate limit status via gh api with
// context support. Returns used and limit counts. Auto-pauses at 75% usage by
// increasing the GitHub rate limit delay.
func (r *RateLimiter) CheckGitHubRateLimitCtx(ctx context.Context) (used, limit int, err error)  /* v090-result-boundary */ {
	if r == nil {
		return 0, 0, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, "gh", "api", "rate_limit", "--jq", ".rate | \"\\(.used) \\(.limit)\"")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, core.E(sonarRatelimitCollectRatelimiterCheckgithubratelimitctx, "gh api rate_limit", err)
	}

	trimmed := core.Trim(string(out))
	parts := textFields(trimmed)
	if len(parts) != 2 {
		return 0, 0, core.E(sonarRatelimitCollectRatelimiterCheckgithubratelimitctx, core.Sprintf("unexpected output %q", trimmed), nil)
	}

	used, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, core.E(sonarRatelimitCollectRatelimiterCheckgithubratelimitctx, "parse used", err)
	}
	limit, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, core.E(sonarRatelimitCollectRatelimiterCheckgithubratelimitctx, "parse limit", err)
	}

	if limit > 0 && float64(used)/float64(limit) >= 0.75 {
		r.SetDelay("github", 5*time.Second)
	}

	return used, limit, nil
}
