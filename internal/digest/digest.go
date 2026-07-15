// Package digest turns a period of the mirrored corpus into clustered candidate
// themes — the mechanical half of `hreysi digest`. It decides WHAT converges;
// a downstream skill decides what to call it.
//
// Convergence is set intersection on explicit per-thread tags, nothing softer. Two
// threads converge iff they share a tag; a connected component of >=2 threads is a
// candidate campaign theme, a lone thread is a post. No model reads free text to
// "find the throughline" — that would be the confabulation the whole pipeline is built
// to avoid.
//
// KNOWN LIMITATION (v1): tag matching is exact-string. Near-synonyms
// (silent-failure vs quiet-failure) do NOT converge. That is the same concept-space
// problem as cross-week reach-back and wants the same (deferred) substrate: extract
// concepts, match by distance, canonicalize. Until then digest makes the fragility
// VISIBLE — it reports singleton tags — rather than letting a near-miss drop silently.
package digest

import (
	"sort"
	"strings"
)

// Thread is one idea from a mirrored entry's threads[] block, with the provenance the
// entry carries.
type Thread struct {
	ID       string
	Thesis   string
	Evidence string
	Section  string
	Angle    string
	Scale    string // post | article | both
	Fit      int
	Tags     []string
	Facts    []string
	Repo     string
	Date     string
}

// HasEvidence reports whether a downstream writer could ground a claim from this
// thread — the anti-fabrication gate. An empty evidence field means "not draftable."
func (t Thread) HasEvidence() bool { return strings.TrimSpace(t.Evidence) != "" }

// ArticleEligible reports whether the thread can anchor long-form (vs only a beat).
func (t Thread) ArticleEligible() bool { return t.Scale == "article" || t.Scale == "both" }

// Cluster is a connected component of threads that share at least one tag — a
// candidate campaign theme.
type Cluster struct {
	Tags    []string // the union of tags across the component (the shared vocabulary)
	Threads []Thread
}

// Repos returns the distinct repos a cluster draws from. Cross-repo convergence is the
// strongest signal a theme is real rather than a single project's preoccupation.
func (c Cluster) Repos() []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range c.Threads {
		if !seen[t.Repo] {
			seen[t.Repo] = true
			out = append(out, t.Repo)
		}
	}
	sort.Strings(out)
	return out
}

// degree counts how many OTHER threads a thread shares a tag with, across the window.
// It is the convergence signal that outranks self-reported fit.
func degree(t Thread, all []Thread) int {
	tset := toSet(t.Tags)
	n := 0
	for _, o := range all {
		if o.ID == t.ID && o.Repo == t.Repo {
			continue
		}
		if intersects(tset, o.Tags) {
			n++
		}
	}
	return n
}

// Cluster partitions threads into shared-tag components. Components of >=2 are campaign
// candidates; singletons are returned separately as posts. Both are sorted by strength:
// clusters by (cross-repo breadth, size, evidence); singles by (fit-adjusted score).
func Clusterize(threads []Thread) (clusters []Cluster, singles []Thread) {
	n := len(threads)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	union := func(a, b int) { parent[find(a)] = find(b) }

	// Edge iff two threads share a tag.
	for i := 0; i < n; i++ {
		si := toSet(threads[i].Tags)
		for j := i + 1; j < n; j++ {
			if intersects(si, threads[j].Tags) {
				union(i, j)
			}
		}
	}

	groups := map[int][]Thread{}
	for i := 0; i < n; i++ {
		r := find(i)
		groups[r] = append(groups[r], threads[i])
	}

	for _, g := range groups {
		if len(g) < 2 {
			singles = append(singles, g...)
			continue
		}
		clusters = append(clusters, Cluster{Tags: unionTags(g), Threads: rankThreads(g, threads)})
	}

	sort.SliceStable(clusters, func(i, j int) bool {
		// Cross-repo breadth first (the strongest signal), then size, then evidence.
		if a, b := len(clusters[i].Repos()), len(clusters[j].Repos()); a != b {
			return a > b
		}
		if a, b := len(clusters[i].Threads), len(clusters[j].Threads); a != b {
			return a > b
		}
		return evidenceCount(clusters[i].Threads) > evidenceCount(clusters[j].Threads)
	})
	singles = rankThreads(singles, threads)
	return clusters, singles
}

// rankThreads orders threads by convergence degree first, then evidence, then fit.
// fit is a self-reported prior and ranks LAST on purpose: a claim with three witnesses
// beats a lonely self-declared 9. That ordering is the anti-inflation lever.
func rankThreads(ts, all []Thread) []Thread {
	out := append([]Thread(nil), ts...)
	sort.SliceStable(out, func(i, j int) bool {
		di, dj := degree(out[i], all), degree(out[j], all)
		if di != dj {
			return di > dj
		}
		if ei, ej := out[i].HasEvidence(), out[j].HasEvidence(); ei != ej {
			return ei
		}
		return out[i].Fit > out[j].Fit
	})
	return out
}

// SingletonTags are tags carried by exactly one thread in the window. Each is either a
// genuinely unique topic or drift (a typo / near-synonym that FAILED to cluster). digest
// surfaces them so the fragility of exact-match tagging is visible, not silent — the hook
// for a future concept-canonicalization layer.
func SingletonTags(threads []Thread) []string {
	count := map[string]int{}
	for _, t := range threads {
		for _, tag := range dedupe(t.Tags) {
			count[tag]++
		}
	}
	var out []string
	for tag, c := range count {
		if c == 1 {
			out = append(out, tag)
		}
	}
	sort.Strings(out)
	return out
}

// Report is the mechanical output: clusters (campaign candidates), singles (posts),
// tag-drift surface, and the coverage note. A naming skill reads this; it never
// recomputes membership.
type Report struct {
	Period    string
	Threads   []Thread
	Clusters  []Cluster
	Singles   []Thread
	Singleton []string
	Coverage  Coverage
}

// Build assembles a report from parsed threads and a coverage note.
func Build(period string, threads []Thread, cov Coverage) Report {
	clusters, singles := Clusterize(threads)
	return Report{
		Period:    period,
		Threads:   threads,
		Clusters:  clusters,
		Singles:   singles,
		Singleton: SingletonTags(threads),
		Coverage:  cov,
	}
}

// ── small set helpers (zero-dep) ──

func toSet(xs []string) map[string]bool {
	s := make(map[string]bool, len(xs))
	for _, x := range xs {
		if x = strings.TrimSpace(x); x != "" {
			s[x] = true
		}
	}
	return s
}

func intersects(set map[string]bool, xs []string) bool {
	for _, x := range xs {
		if set[strings.TrimSpace(x)] {
			return true
		}
	}
	return false
}

func unionTags(ts []Thread) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range ts {
		for _, tag := range t.Tags {
			if tag = strings.TrimSpace(tag); tag != "" && !seen[tag] {
				seen[tag] = true
				out = append(out, tag)
			}
		}
	}
	sort.Strings(out)
	return out
}

func dedupe(xs []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range xs {
		x = strings.TrimSpace(x)
		if x != "" && !seen[x] {
			seen[x] = true
			out = append(out, x)
		}
	}
	return out
}

func evidenceCount(ts []Thread) int {
	n := 0
	for _, t := range ts {
		if t.HasEvidence() {
			n++
		}
	}
	return n
}
