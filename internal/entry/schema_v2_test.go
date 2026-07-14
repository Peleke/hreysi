package entry

import "strings"

import "testing"

// The v2 narrative schema puts frontmatter + a thesis title + topical sections
// ABOVE the mechanical spine, and moves "## Commits" to the BOTTOM of the file.
// Capture must keep working against that layout: a later commit on the same day
// has to land inside the Commits section without touching frontmatter or prose.
const v2Entry = `---
title: "The config that was never read"
date: 2026-07-14
project: hreysi
tags: [mcp, config]
status: shipped
threads:
  - id: dead-mcp-config
    thesis: "settings.json mcpServers was never read"
    arcRole: field_note
---

# The config that was never read

## The gap we closed

Two MCP servers had never once loaded.

## The Journey

It looked like duplication. It was worse than duplication.

## Improvements

### Gotchas
- Dual config files silently diverge.

## Commits

### ` + "`aaa1111`" + ` — fix: drop dead mcpServers block
_2026-07-14T10:00:00-04:00_

Files:
- ` + "`settings.json`" + `
`

func TestInsertCommit_V2SchemaCommitsLast(t *testing.T) {
	block := "\n### `bbb2222` — feat: mirror\n_2026-07-14T11:00:00-04:00_\n\n"
	got := insertCommit(v2Entry, block)

	// 1. Frontmatter must survive intact and stay at the very top.
	if !strings.HasPrefix(got, "---\ntitle: \"The config that was never read\"") {
		t.Fatalf("frontmatter corrupted or displaced; head=%q", got[:min(80, len(got))])
	}
	// 2. The machine-readable seam must survive.
	if !strings.Contains(got, "threads:") || !strings.Contains(got, "arcRole: field_note") {
		t.Error("threads[] seam was damaged")
	}
	// 3. Narrative prose must survive.
	if !strings.Contains(got, "It looked like duplication. It was worse than duplication.") {
		t.Error("narrative prose was damaged")
	}
	// 4. The new commit must land AFTER the pre-existing one (inside ## Commits).
	iOld := strings.Index(got, "aaa1111")
	iNew := strings.Index(got, "bbb2222")
	if iNew < 0 {
		t.Fatal("new commit block was not inserted at all")
	}
	if iNew < iOld {
		t.Error("new commit landed before the existing one")
	}
	// 5. The new commit must NOT be injected into the narrative sections.
	iJourney := strings.Index(got, "## The Journey")
	iCommits := strings.Index(got, "## Commits")
	if iNew < iCommits {
		t.Errorf("commit block leaked ABOVE the ## Commits section (into prose)")
	}
	if iJourney > iCommits {
		t.Error("section order inverted; Commits should remain last")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
