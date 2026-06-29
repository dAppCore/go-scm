// SPDX-License-Identifier: EUPL-1.2

package handlers

import (
	core "dappco.re/go"
	"dappco.re/go/scm/jobrunner"
)

func TestHandlers_completionComment_Good(t *core.T) {
	got := completionComment(&jobrunner.PipelineSignal{PRNumber: 7, Message: "all green"})
	core.AssertContains(t, got, "#7")
	core.AssertContains(t, got, "all green")
}

func TestHandlers_completionComment_Bad(t *core.T) {
	got := completionComment(&jobrunner.PipelineSignal{})
	core.AssertContains(t, got, "#0")
}

func TestHandlers_completionComment_Ugly(t *core.T) {
	got := completionComment(&jobrunner.PipelineSignal{PRNumber: -1})
	core.AssertContains(t, got, "#-1")
}

func TestHandlers_fixCommandComment_Good(t *core.T) {
	got := fixCommandComment(&jobrunner.PipelineSignal{Mergeable: "conflicting"})
	core.AssertContains(t, got, "merge conflicts")
}

func TestHandlers_fixCommandComment_Bad(t *core.T) {
	got := fixCommandComment(&jobrunner.PipelineSignal{ThreadsTotal: 2, ThreadsResolved: 0})
	core.AssertContains(t, got, "unresolved review threads")
}

func TestHandlers_fixCommandComment_Ugly(t *core.T) {
	got := fixCommandComment(&jobrunner.PipelineSignal{})
	core.AssertContains(t, got, "requested fixes")
}

func TestHandlers_tickCheckbox_Good(t *core.T) {
	body := "- [ ] task #3\n- [ ] task #4\n"
	got, changed := tickCheckbox(body, 3)
	core.AssertTrue(t, changed)
	core.AssertContains(t, got, "[x] task #3")
	core.AssertContains(t, got, "[ ] task #4")
}

func TestHandlers_tickCheckbox_Bad(t *core.T) {
	body := "- [ ] task #3\n"
	got, changed := tickCheckbox(body, 99)
	core.AssertFalse(t, changed)
	core.AssertContains(t, got, "[ ] task #3")
}

func TestHandlers_tickCheckbox_Ugly(t *core.T) {
	body := "- [X] task #3\n"
	got, changed := tickCheckbox(body, 3)
	core.AssertTrue(t, changed)
	core.AssertContains(t, got, "[X] task #3")
}
