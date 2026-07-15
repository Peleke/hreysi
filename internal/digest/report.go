package digest

import (
	"fmt"
	"strings"
)

// Markdown renders a Digest Report for a human to read and for the naming skill to
// consume. It presents SELECTION only — clusters, singles, drift, coverage — and never
// names a theme or writes prose. That is the skill's job; the report hands it the
// membership it may not change.
func (r Report) Markdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Digest — %s\n\n", r.Period)
	fmt.Fprintf(&b, "%d threads across %d repos → %d campaign candidate(s), %d standalone post(s).\n\n",
		len(r.Threads), len(r.Coverage.Threaded), len(r.Clusters), len(r.Singles))

	if len(r.Clusters) == 0 {
		b.WriteString("_No convergence this period — nothing shares a tag with anything else. ")
		b.WriteString("Every thread is a standalone post._\n\n")
	}

	for i, c := range r.Clusters {
		fmt.Fprintf(&b, "## Campaign candidate %d — tags: %s\n", i+1, strings.Join(c.Tags, ", "))
		repos := c.Repos()
		note := ""
		if len(repos) > 1 {
			note = "  ← **cross-repo** (strongest signal)"
		}
		fmt.Fprintf(&b, "%d threads across %s%s\n\n", len(c.Threads), strings.Join(repos, ", "), note)
		for _, t := range c.Threads {
			fmt.Fprintf(&b, "- **%s** (%s, %s, fit %d) — %s\n", t.ID, t.Repo, t.Scale, t.Fit, t.Thesis)
			if !t.HasEvidence() {
				b.WriteString("    ⚠ no evidence — not draftable as-is\n")
			}
		}
		b.WriteString("\n")
	}

	if len(r.Singles) > 0 {
		b.WriteString("## Standalone posts (no convergence)\n\n")
		for _, t := range r.Singles {
			fmt.Fprintf(&b, "- **%s** (%s, %s, fit %d) — %s\n", t.ID, t.Repo, t.Scale, t.Fit, t.Thesis)
		}
		b.WriteString("\n")
	}

	if len(r.Singleton) > 0 {
		b.WriteString("## ⚠ tag drift — used once, clustered nothing\n\n")
		b.WriteString("Exact-match tagging. These tags appear on exactly one thread — either genuinely\n")
		b.WriteString("unique, or a near-synonym that failed to converge (e.g. `silent-failure` vs\n")
		b.WriteString("`quiet-failure`). A concept-canonicalization layer would merge the latter; until\n")
		b.WriteString("then, they are surfaced here rather than dropped silently:\n\n")
		fmt.Fprintf(&b, "  %s\n\n", strings.Join(r.Singleton, ", "))
	}

	b.WriteString("## Coverage\n\n")
	fmt.Fprintf(&b, "- captured + threaded this period: %s\n",
		orNone(strings.Join(r.Coverage.Threaded, ", ")))
	if len(r.Coverage.Uncovered) > 0 {
		fmt.Fprintf(&b, "- ⚠ session activity, NOT threaded: **%s**\n", strings.Join(r.Coverage.Uncovered, ", "))
		b.WriteString("  → cross-project themes involving these repos are INVISIBLE to digest.\n")
		b.WriteString("    run capture + expand + mirror there to make them minable.\n")
	} else {
		b.WriteString("- no un-threaded repo activity detected in-window (best-effort)\n")
	}
	return b.String()
}

func orNone(s string) string {
	if s == "" {
		return "_none_"
	}
	return s
}
