// SPDX-License-Identifier: EUPL-1.2

package collect

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"sync"
	"time"
)

// RateLimiter tracks per-source rate limiting to avoid overwhelming APIs.
type RateLimiter struct {
	mu     sync.Mutex
	delays map[string]time.Duration
	last   map[string]time.Time
}

// NewRateLimiter creates a limiter with default delays.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		delays: map[string]time.Duration{
			"github": 5 * time.Second,
		},
		last: make(map[string]time.Time),
	}
}

// GetDelay returns the delay configured for a source.
func (r *RateLimiter) GetDelay(source string) time.Duration {
	if r == nil {
		return 0
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.delays[source]
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

// Wait blocks until the rate limit allows the next request.
func (r *RateLimiter) Wait(ctx context.Context, source string) error {
	if r == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.mu.Lock()
	delay := r.delays[source]
	last := r.last[source]
	now := time.Now()
	if delay <= 0 {
		r.last[source] = now
		r.mu.Unlock()
		return nil
	}
	wait := last.Add(delay).Sub(now)
	if wait < 0 {
		wait = 0
	}
	r.mu.Unlock()

	timer := time.NewTimer(wait)
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

// CheckGitHubRateLimit checks GitHub API rate limit status via gh api.
func (r *RateLimiter) CheckGitHubRateLimit() (used, limit int, err error) {
	return r.CheckGitHubRateLimitCtx(context.Background())
}

// CheckGitHubRateLimitCtx checks GitHub API rate limit status via gh api with context support.
func (r *RateLimiter) CheckGitHubRateLimitCtx(ctx context.Context) (used, limit int, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "gh", "api", "rate_limit")
	out, runErr := cmd.Output()
	if runErr != nil {
		return 0, 0, runErr
	}
	var payload struct {
		Resources struct {
			Core struct {
				Limit int `json:"limit"`
				Used  int `json:"used"`
			} `json:"core"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(out, &payload); err != nil {
		return 0, 0, err
	}
	if payload.Resources.Core.Limit == 0 {
		return 0, 0, errors.New("collect: gh api returned zero rate limit")
	}
	return payload.Resources.Core.Used, payload.Resources.Core.Limit, nil
}
