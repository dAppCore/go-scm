// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	fmt "dappco.re/go/core/scm/internal/ax/fmtx"
	strings "dappco.re/go/core/scm/internal/ax/stringsx"
	exec "golang.org/x/sys/execabs"
	"maps"
	"strconv"
	"sync"
	"time"

	core "dappco.re/go/core/log"
)

// RateLimiter tracks per-source rate limiting to avoid overwhelming APIs.
type RateLimiter struct {
	mu     sync.Mutex
	delays map[string]time.Duration
	last   map[string]time.Time
}

// Default rate limit delays per source.
var defaultDelays = map[string]time.Duration{
	"github":      500 * time.Millisecond,
	"bitcointalk": 2 * time.Second,
	"coingecko":   1500 * time.Millisecond,
	"iacr":        1 * time.Second,
	"arxiv":       1 * time.Second,
}

// NewRateLimiter creates a limiter with default delays.
// Usage: NewRateLimiter(...)
func NewRateLimiter() *RateLimiter {
	delays := make(map[string]time.Duration, len(defaultDelays))
	maps.Copy(delays, defaultDelays)
	return &RateLimiter{
		delays: delays,
		last:   make(map[string]time.Time),
	}
}

// Wait blocks until the rate limit allows the next request for the given source.
// It respects context cancellation.
// Usage: Wait(...)
func (r *RateLimiter) Wait(ctx context.Context, source string) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	if r.delays == nil {
		r.delays = make(map[string]time.Duration, len(defaultDelays))
		maps.Copy(r.delays, defaultDelays)
	}
	if r.last == nil {
		r.last = make(map[string]time.Time)
	}
	delay, ok := r.delays[source]
	if !ok {
		delay = 500 * time.Millisecond
	}
	lastTime := r.last[source]

	elapsed := time.Since(lastTime)
	if elapsed >= delay {
		// Enough time has passed — claim the slot immediately.
		r.last[source] = time.Now()
		r.mu.Unlock()
		return nil
	}

	remaining := delay - elapsed
	r.mu.Unlock()

	// Wait outside the lock, then reclaim.
	select {
	case <-ctx.Done():
		return core.E("collect.RateLimiter.Wait", "context cancelled", ctx.Err())
	case <-time.After(remaining):
	}

	r.mu.Lock()
	r.last[source] = time.Now()
	r.mu.Unlock()

	return nil
}

// SetDelay sets the delay for a source.
// Usage: SetDelay(...)
func (r *RateLimiter) SetDelay(source string, d time.Duration) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.delays == nil {
		r.delays = make(map[string]time.Duration, len(defaultDelays))
		maps.Copy(r.delays, defaultDelays)
	}
	r.delays[source] = d
}

// GetDelay returns the delay configured for a source.
// Usage: GetDelay(...)
func (r *RateLimiter) GetDelay(source string) time.Duration {
	if r == nil {
		return 500 * time.Millisecond
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.delays == nil {
		r.delays = make(map[string]time.Duration, len(defaultDelays))
		maps.Copy(r.delays, defaultDelays)
	}
	if d, ok := r.delays[source]; ok {
		return d
	}
	return 500 * time.Millisecond
}

// CheckGitHubRateLimit checks GitHub API rate limit status via gh api.
// Returns used and limit counts. Auto-pauses at 75% usage by increasing
// the GitHub rate limit delay.
// Deprecated: Use CheckGitHubRateLimitCtx for context-aware cancellation.
// Usage: CheckGitHubRateLimit(...)
func (r *RateLimiter) CheckGitHubRateLimit() (used, limit int, err error) {
	return r.CheckGitHubRateLimitCtx(context.Background())
}

// CheckGitHubRateLimitCtx checks GitHub API rate limit status via gh api with context support.
// Returns used and limit counts. Auto-pauses at 75% usage by increasing
// the GitHub rate limit delay.
// Usage: CheckGitHubRateLimitCtx(...)
func (r *RateLimiter) CheckGitHubRateLimitCtx(ctx context.Context) (used, limit int, err error) {
	if r == nil {
		r = NewRateLimiter()
	}
	cmd := exec.CommandContext(ctx, "gh", "api", "rate_limit", "--jq", ".rate | \"\\(.used) \\(.limit)\"")
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, core.E("collect.RateLimiter.CheckGitHubRateLimit", "failed to check rate limit", err)
	}

	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return 0, 0, core.E("collect.RateLimiter.CheckGitHubRateLimit",
			fmt.Sprintf("unexpected output format: %q", string(out)), nil)
	}

	used, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, core.E("collect.RateLimiter.CheckGitHubRateLimit", "failed to parse used count", err)
	}

	limit, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, core.E("collect.RateLimiter.CheckGitHubRateLimit", "failed to parse limit count", err)
	}

	// Auto-pause at 75% usage
	if limit > 0 {
		usage := float64(used) / float64(limit)
		if usage >= 0.75 {
			r.SetDelay("github", 5*time.Second)
		}
	}

	return used, limit, nil
}
