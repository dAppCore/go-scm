// SPDX-License-Identifier: EUPL-1.2

package forgejo

import (
	"context"

	coreforge "dappco.re/go/scm/forge"
	"dappco.re/go/scm/jobrunner"
)

type Config struct {
	Repos []string
}

type ForgejoSource struct {
	repos []string
	forge *coreforge.Client
}

func New(cfg Config, client *coreforge.Client) *ForgejoSource {
	return &ForgejoSource{repos: append([]string(nil), cfg.Repos...), forge: client}
}

func (s *ForgejoSource) Name() string { return "forgejo" }

func (s *ForgejoSource) Poll(ctx context.Context) ([]*jobrunner.PipelineSignal, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (s *ForgejoSource) Report(ctx context.Context, result *jobrunner.ActionResult) error {
	if ctx != nil {
		return ctx.Err()
	}
	_ = result
	return nil
}
