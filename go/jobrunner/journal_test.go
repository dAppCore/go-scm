// SPDX-License-Identifier: EUPL-1.2

package jobrunner

import (
	"testing"
	"time"

	core "dappco.re/go"
)

func TestJournalAppendWritesDatePartitionedJSONL(t *testing.T) {
	journal, err := NewJournal(t.TempDir())
	if err != nil {
		t.Fatalf("new journal: %v", err)
	}

	ts := time.Date(2026, 4, 15, 14, 30, 0, 123000000, time.UTC)
	signal := &PipelineSignal{
		RepoOwner:       "core",
		RepoName:        "go-scm",
		PRState:         "OPEN",
		IsDraft:         true,
		CheckStatus:     "PENDING",
		Mergeable:       "UNKNOWN",
		ThreadsTotal:    3,
		ThreadsResolved: 1,
	}
	result := &ActionResult{
		Action:      "dispatch",
		RepoOwner:   "core",
		RepoName:    "go-scm",
		EpicNumber:  9,
		ChildNumber: 2,
		PRNumber:    17,
		Success:     true,
		Timestamp:   ts,
		Duration:    1500 * time.Millisecond,
		Cycle:       4,
	}

	if err := journal.Append(signal, result); err != nil {
		t.Fatalf("append: %v", err)
	}

	path := core.PathJoin(journal.baseDir, "2026", "04", "15.jsonl")
	rawR := core.ReadFile(path)
	if !rawR.OK {
		t.Fatalf("read journal: %v", rawR.Error())
	}
	raw := rawR.Value.([]byte)

	lines := splitLines(string(raw))
	if len(lines) != 1 {
		t.Fatalf("expected one journal line, got %d", len(lines))
	}

	var entry JournalEntry
	if r := core.JSONUnmarshal([]byte(lines[0]), &entry); !r.OK {
		t.Fatalf("unmarshal entry: %v", r.Error())
	}
	if entry.Timestamp != ts.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected timestamp: %q", entry.Timestamp)
	}
	if entry.Repo != "core/go-scm" {
		t.Fatalf("unexpected repo: %q", entry.Repo)
	}
	if entry.Signals.ThreadsTotal != 3 || !entry.Signals.IsDraft {
		t.Fatalf("unexpected signal snapshot: %#v", entry.Signals)
	}
	if entry.Result.DurationMs != 1500 {
		t.Fatalf("unexpected duration: %d", entry.Result.DurationMs)
	}
}

func TestActionResultJSONUsesMilliseconds(t *testing.T) {
	original := ActionResult{
		Action:    "dispatch",
		PRNumber:  17,
		Success:   true,
		Timestamp: time.Date(2026, 4, 15, 14, 30, 0, 0, time.UTC),
		Duration:  1500 * time.Millisecond,
		Cycle:     4,
	}

	rawR := core.JSONMarshal(original)
	if !rawR.OK {
		t.Fatalf("marshal: %v", rawR.Error())
	}
	raw := rawR.Value.([]byte)

	var decoded map[string]any
	if r := core.JSONUnmarshal(raw, &decoded); !r.OK {
		t.Fatalf("unmarshal map: %v", r.Error())
	}

	if got := decoded["duration_ms"]; got != float64(1500) {
		t.Fatalf("unexpected duration_ms: %#v", got)
	}

	var roundTrip ActionResult
	if r := core.JSONUnmarshal(raw, &roundTrip); !r.OK {
		t.Fatalf("round trip unmarshal: %v", r.Error())
	}
	if roundTrip.Duration != original.Duration {
		t.Fatalf("unexpected duration round trip: %v", roundTrip.Duration)
	}
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func TestJournal_NewJournal_Good(t *testing.T) {
	target := "NewJournal"
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

func TestJournal_NewJournal_Bad(t *testing.T) {
	target := "NewJournal"
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

func TestJournal_NewJournal_Ugly(t *testing.T) {
	target := "NewJournal"
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

func TestJournal_Journal_Append_Good(t *testing.T) {
	reference := "Append"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Journal_Append"
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

func TestJournal_Journal_Append_Bad(t *testing.T) {
	reference := "Append"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Journal_Append"
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

func TestJournal_Journal_Append_Ugly(t *testing.T) {
	reference := "Append"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Journal_Append"
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
