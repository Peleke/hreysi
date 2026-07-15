package digest

import (
	"os"
	"path/filepath"
	"testing"
)

// A trimmed real-shape entry: three threads, two sharing a tag, one standalone.
const entry = `---
title: "The tool that lied"
date: 2026-07-14
project: hreysi
tags: [dogfood, doctor]
threads:
  - id: doctor-false-green
    thesis: "doctor reported green while capture was dead"
    evidence: "checked hook text, not target"
    section: "## A"
    tags: [silent-failure, health-check]
    angle: field_note
    scale: both
    fit: 9
    facts:
      - "hreysi 1.1.1"
      - "the hook target was /tmp/hreysi-build/hreysi"
  - id: verdict-subset
    thesis: "the verdict restated a subset of its checks"
    evidence: "Healthy omitted the target check"
    section: "## A"
    tags: [silent-failure]
    angle: demystifier
    scale: post
    fit: 7
    facts: []
  - id: mirror-guard
    thesis: "ownership is a key, not a substring"
    evidence: "a note mentioning the marker was clobbered"
    section: "## B"
    tags: [provenance, data-safety]
    angle: demystifier
    scale: both
    fit: 8
    facts: []
---

# body`

func TestParseEntry_ReadsAllFieldsAndArrayStyles(t *testing.T) {
	ts := ParseEntry(entry, "hreysi", "2026-07-14")
	if len(ts) != 3 {
		t.Fatalf("expected 3 threads, got %d", len(ts))
	}
	a := ts[0]
	if a.ID != "doctor-false-green" {
		t.Errorf("id = %q", a.ID)
	}
	if a.Fit != 9 {
		t.Errorf("fit = %d, want 9", a.Fit)
	}
	if a.Scale != "both" || a.Angle != "field_note" {
		t.Errorf("scale/angle = %q/%q", a.Scale, a.Angle)
	}
	// inline array
	if len(a.Tags) != 2 || a.Tags[0] != "silent-failure" {
		t.Errorf("tags = %v", a.Tags)
	}
	// block-list array, including a value containing a slash and no quotes issues
	if len(a.Facts) != 2 || a.Facts[0] != "hreysi 1.1.1" {
		t.Errorf("facts = %v", a.Facts)
	}
	if a.Repo != "hreysi" || a.Date != "2026-07-14" {
		t.Errorf("provenance = %q/%q", a.Repo, a.Date)
	}
	// empty inline array
	if len(ts[1].Facts) != 0 {
		t.Errorf("verdict-subset facts = %v, want empty", ts[1].Facts)
	}
}

func TestCluster_SharedTagConvergesTwoThreadsLoneStandsAlone(t *testing.T) {
	ts := ParseEntry(entry, "hreysi", "2026-07-14")
	clusters, singles := Clusterize(ts)

	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster (silent-failure), got %d", len(clusters))
	}
	if len(clusters[0].Threads) != 2 {
		t.Errorf("cluster size = %d, want 2", len(clusters[0].Threads))
	}
	if len(singles) != 1 || singles[0].ID != "mirror-guard" {
		t.Errorf("expected mirror-guard standalone, got %+v", singles)
	}
}

func TestCluster_RanksByConvergenceNotFit(t *testing.T) {
	// Two threads share a tag (degree 1 each); a third has a higher fit but no shared
	// tag. Within the cluster, the convergent threads must outrank on degree; the lone
	// high-fit thread is a single, not smuggled into the cluster.
	ts := ParseEntry(entry, "hreysi", "2026-07-14")
	clusters, singles := Clusterize(ts)
	// mirror-guard has fit 8 (higher than verdict-subset's 7) but is a SINGLE, because
	// fit does not create convergence — only shared tags do.
	for _, c := range clusters {
		for _, th := range c.Threads {
			if th.ID == "mirror-guard" {
				t.Error("fit pulled a non-converging thread into a cluster")
			}
		}
	}
	if len(singles) == 0 || singles[0].ID != "mirror-guard" {
		t.Errorf("high-fit lone thread should be the top single, got %+v", singles)
	}
}

func TestSingletonTags_SurfacesDrift(t *testing.T) {
	ts := ParseEntry(entry, "hreysi", "2026-07-14")
	singles := SingletonTags(ts)
	// silent-failure appears twice -> not a singleton. health-check, provenance,
	// data-safety each once -> singletons (candidates for drift / a future concept merge).
	for _, s := range singles {
		if s == "silent-failure" {
			t.Error("silent-failure clustered (used twice) — must not be a singleton")
		}
	}
	want := map[string]bool{"health-check": true, "provenance": true, "data-safety": true}
	for _, s := range singles {
		if !want[s] {
			t.Errorf("unexpected singleton %q", s)
		}
	}
	if len(singles) != 3 {
		t.Errorf("expected 3 singleton tags, got %v", singles)
	}
}

// Cross-repo convergence is the strongest signal; a cluster spanning two repos must
// outrank a bigger single-repo cluster.
func TestCluster_CrossRepoRanksFirst(t *testing.T) {
	threads := []Thread{
		{ID: "a", Tags: []string{"x"}, Repo: "r1", Evidence: "e"},
		{ID: "b", Tags: []string{"x"}, Repo: "r1", Evidence: "e"},
		{ID: "c", Tags: []string{"x"}, Repo: "r1", Evidence: "e"}, // 3 threads, 1 repo
		{ID: "d", Tags: []string{"y"}, Repo: "r1", Evidence: "e"},
		{ID: "e", Tags: []string{"y"}, Repo: "r2", Evidence: "e"}, // 2 threads, 2 repos
	}
	clusters, _ := Clusterize(threads)
	if len(clusters) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(clusters))
	}
	if len(clusters[0].Repos()) != 2 {
		t.Errorf("cross-repo cluster should rank first; got tags %v across %v", clusters[0].Tags, clusters[0].Repos())
	}
}

func TestParseCorpus_WindowAndLegacySkip(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("2026-07-14-hreysi.md", entry)
	write("2026-07-01-hreysi.md", entry)                                     // out of window
	write("2026-03-15-cadence-agent-stack.md", "# legacy\nno frontmatter\n") // legacy, no threads

	threads, repos, err := ParseCorpus(dir, "2026-07-10", "2026-07-20")
	if err != nil {
		t.Fatal(err)
	}
	if len(threads) != 3 {
		t.Errorf("expected 3 threads from the in-window entry, got %d", len(threads))
	}
	if !repos["hreysi"] || len(repos) != 1 {
		t.Errorf("repos = %v, want just hreysi", repos)
	}
}

func TestDetectCoverage_NamesUncoveredRepos(t *testing.T) {
	projects := t.TempDir()
	// A repo with a recent local buildlog entry but NOT in the threaded set.
	linwheel := filepath.Join(projects, "linwheel", "buildlog")
	if err := os.MkdirAll(linwheel, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(linwheel, "2026-07-14.md"), []byte("# x"), 0o644); err != nil {
		t.Fatal(err)
	}

	cov := DetectCoverage(projects, "2026-07-10", "2026-07-20", map[string]bool{"hreysi": true})
	if len(cov.Uncovered) != 1 || cov.Uncovered[0] != "linwheel" {
		t.Errorf("expected linwheel uncovered, got %+v", cov)
	}
}
