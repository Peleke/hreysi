package campaign

import (
	"os"
	"path/filepath"
	"testing"
)

const draftBrief = `---
type: campaign-brief
period: 2026-W29
theme: "The silent failure"
status: draft
articles:
  - id: a1
    kind: thought
    thesis: "Silent success is the signature failure of well-tested code"
    editorial_map:
      - claim: "a health check that greps a marker reports green over a dead system"
        evidence: "checked hook text not target"
  - id: t1
    kind: tutorial
    pairs_with: a1
    capability: "Write a git hook whose failure is impossible to miss"
beats:
  - order: 1
    id: should-not-be-parsed-as-article
---

## Notes
body`

func write(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseBrief_ReadsHeadAndArticlesNotBeats(t *testing.T) {
	dir := t.TempDir()
	p := write(t, dir, "2026-W29.md", draftBrief)
	b, ok, err := ParseBrief(p)
	if err != nil || !ok {
		t.Fatalf("parse failed: ok=%v err=%v", ok, err)
	}
	if b.Status != "draft" || b.Period != "2026-W29" {
		t.Errorf("head = %q / %q", b.Status, b.Period)
	}
	if len(b.Articles) != 2 {
		t.Fatalf("expected 2 articles (not the beat), got %d: %+v", len(b.Articles), b.Articles)
	}
	if b.Articles[0].ID != "a1" || b.Articles[0].Kind != "thought" {
		t.Errorf("a1 = %+v", b.Articles[0])
	}
	if b.Articles[1].ID != "t1" || b.Articles[1].Kind != "tutorial" {
		t.Errorf("t1 = %+v", b.Articles[1])
	}
	// The beat's id must NOT be parsed as an article.
	for _, a := range b.Articles {
		if a.ID == "should-not-be-parsed-as-article" {
			t.Error("a beat leaked into articles[]")
		}
	}
}

// The gate is the whole point: a draft brief blocks a run.
func TestBuildPlan_DraftBriefIsBlocked(t *testing.T) {
	dir := t.TempDir()
	b, _, _ := ParseBrief(write(t, dir, "b.md", draftBrief))
	plan := BuildPlan(b, t.TempDir())
	if plan.Blocked == "" {
		t.Fatal("a draft brief must block a run")
	}
	if len(plan.WorkList) != 0 {
		t.Error("a blocked plan must have an empty work-list — no tokens would be spent")
	}
}

func TestBuildPlan_ApprovedBriefWorkListsUndrafted(t *testing.T) {
	dir := t.TempDir()
	approved := draftBrief
	approved = replace(approved, "status: draft", "status: approved\ndrafted: [a1]")
	b, ok, err := ParseBrief(write(t, dir, "b.md", approved))
	if err != nil || !ok {
		t.Fatalf("parse: ok=%v err=%v", ok, err)
	}
	if !b.Approved() {
		t.Fatal("brief should be approved")
	}
	plan := BuildPlan(b, t.TempDir())
	if plan.Blocked != "" {
		t.Fatalf("approved brief should not be blocked: %s", plan.Blocked)
	}
	// a1 is in drafted:; only t1 should be work-listed. Idempotency = no re-spend.
	if len(plan.WorkList) != 1 || plan.WorkList[0].ID != "t1" {
		t.Errorf("work-list = %+v, want [t1] only (a1 already drafted)", plan.WorkList)
	}
	if len(plan.Skipped) != 1 || plan.Skipped[0] != "a1" {
		t.Errorf("skipped = %v, want [a1]", plan.Skipped)
	}
}

func TestBuildPlan_FullyDrafted(t *testing.T) {
	dir := t.TempDir()
	approved := replace(draftBrief, "status: draft", "status: approved\ndrafted: [a1, t1]")
	b, _, _ := ParseBrief(write(t, dir, "b.md", approved))
	plan := BuildPlan(b, t.TempDir())
	if len(plan.WorkList) != 0 {
		t.Errorf("a fully-drafted brief must work-list nothing (re-run spends zero), got %+v", plan.WorkList)
	}
}

func TestNextPortfolioNumber(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "01-a.md", "x")
	write(t, dir, "09-b.md", "x")
	write(t, dir, "_ideas.md", "x") // non-numbered, ignored
	if got := nextPortfolioNumber(dir); got != 10 {
		t.Errorf("next number = %d, want 10", got)
	}
	if got := nextPortfolioNumber(t.TempDir()); got != 1 {
		t.Errorf("empty portfolio next = %d, want 1", got)
	}
}

func TestParseBrief_NonBriefReturnsNotOk(t *testing.T) {
	dir := t.TempDir()
	p := write(t, dir, "note.md", "---\ntitle: just a note\n---\nbody")
	_, ok, err := ParseBrief(p)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("a non-brief file must return ok=false, so a Campaigns dir can hold other notes")
	}
}

// tiny local string replace to avoid importing strings in the test just for one call
func replace(s, old, new string) string {
	i := indexOf(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + new + s[i+len(old):]
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
