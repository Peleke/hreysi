// Package campaign is the mechanical half of `hreysi campaign` — the orchestrator that
// fires an approved Campaign Brief's article drafts. This is P1: parse a brief, gate on
// approval, compute the work-list, plan output numbers. It spawns NO agents and spends no
// tokens; drafting (P2+) is agentic and lands later. See docs/campaign-run-spec.md.
//
// The gate and the drafted-set idempotency live here, in compiled code, on purpose: they
// are what protect article-draft spend, and an if-statement cannot be prompted around.
package campaign

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Article is one entry from a brief's articles[]. Only the fields P1 needs to gate and
// plan are parsed; the writers read the rest of the brief directly when they draft.
type Article struct {
	ID   string
	Kind string // thought | tutorial
	// Thesis for thought, Capability for tutorial — used only for the human-facing plan.
	Summary string
}

// Brief is the parsed head of a Campaign Brief.
type Brief struct {
	Path     string
	Period   string
	Theme    string
	Status   string   // draft | approved
	Drafted  []string // article ids already drafted (idempotency set)
	Articles []Article
}

// Approved reports whether the brief has cleared the human token gate.
func (b Brief) Approved() bool { return strings.TrimSpace(b.Status) == "approved" }

// InDrafted reports whether an article id is already in a brief's drafted set.
func InDrafted(b Brief, id string) bool { return inSet(b.Drafted, id) }

// WorkList is the article briefs not yet drafted — what a run would spend tokens on.
func (b Brief) WorkList() []Article {
	done := map[string]bool{}
	for _, d := range b.Drafted {
		done[strings.TrimSpace(d)] = true
	}
	var out []Article
	for _, a := range b.Articles {
		if !done[a.ID] {
			out = append(out, a)
		}
	}
	return out
}

// Plan is the mechanical decision P1 produces: whether a run may proceed, what it would
// draft, and where. No side effects, no tokens.
type Plan struct {
	Brief    Brief
	Blocked  string    // non-empty if a run must not proceed (e.g. not approved)
	WorkList []Article // articles a run would draft
	Skipped  []string  // article ids already in drafted:
	NextNum  int       // the next NN- for portfolio output
}

// BuildPlan gates the brief and computes the work-list against the portfolio's numbering.
func BuildPlan(b Brief, portfolioDir string) Plan {
	p := Plan{Brief: b, NextNum: nextPortfolioNumber(portfolioDir)}
	for _, a := range b.Articles {
		if inSet(b.Drafted, a.ID) {
			p.Skipped = append(p.Skipped, a.ID)
		}
	}
	if !b.Approved() {
		p.Blocked = fmt.Sprintf("brief is status:%s — review and set status:approved first (this is the token gate)", orUnset(b.Status))
		return p
	}
	p.WorkList = b.WorkList()
	return p
}

// ── parsing (schema-specific, zero-dep — hreysi owns the brief schema) ──

var (
	reType    = regexp.MustCompile(`(?m)^type:\s*campaign-brief\s*$`)
	reScalar  = func(k string) *regexp.Regexp { return regexp.MustCompile(`(?m)^` + k + `:\s*"?(.*?)"?\s*$`) }
	reArtID   = regexp.MustCompile(`(?m)^\s*-\s+id:\s*(\w+)`)
	reArtKind = regexp.MustCompile(`(?m)^\s*kind:\s*(\w+)`)
)

// ParseBrief reads a Campaign Brief's frontmatter. It is intentionally forgiving: a file
// that is not a campaign brief returns ok=false rather than an error, so a Campaigns/
// directory can hold other notes.
func ParseBrief(path string) (Brief, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Brief{}, false, err
	}
	content := string(data)
	fm := frontmatter(content)
	if fm == "" || !reType.MatchString(fm) {
		return Brief{}, false, nil
	}
	b := Brief{Path: path}
	b.Period = scalar(fm, "period")
	b.Theme = scalar(fm, "theme")
	b.Status = scalar(fm, "status")
	b.Drafted = parseInlineList(fm, "drafted")
	b.Articles = parseArticles(fm)
	return b, true, nil
}

func frontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return ""
	}
	return rest[:end]
}

func scalar(fm, key string) string {
	m := reScalar(key).FindStringSubmatch(fm)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(m[1])
}

// parseInlineList reads `key: [a, b, c]` (the drafted: idempotency set). Absent or `[]`
// yields nil.
func parseInlineList(fm, key string) []string {
	re := regexp.MustCompile(`(?m)^` + key + `:\s*\[(.*?)\]\s*$`)
	m := re.FindStringSubmatch(fm)
	if m == nil {
		return nil
	}
	var out []string
	for _, p := range strings.Split(m[1], ",") {
		if p = strings.TrimSpace(strings.Trim(p, `"'`)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseArticles reads the articles[] list — id + kind + a one-line summary for the plan.
// Only the top-level articles: block is scanned (not beats:), keyed on `- id:` items.
func parseArticles(fm string) []Article {
	block := listBlock(fm, "articles:")
	if block == "" {
		return nil
	}
	var out []Article
	for _, item := range splitItems(block) {
		id := reArtID.FindStringSubmatch(item)
		if id == nil {
			continue
		}
		a := Article{ID: id[1]}
		if k := reArtKind.FindStringSubmatch(item); k != nil {
			a.Kind = k[1]
		}
		// Summary: thesis (thought) or capability (tutorial), whichever is present.
		if s := scalar(item, `\s*thesis`); s != "" {
			a.Summary = s
		} else if s := scalar(item, `\s*capability`); s != "" {
			a.Summary = s
		}
		out = append(out, a)
	}
	return out
}

// listBlock returns the indented lines under a top-level `key:` up to the next top-level key.
func listBlock(fm, key string) string {
	lines := strings.Split(fm, "\n")
	start := -1
	for i, l := range lines {
		if strings.TrimSpace(l) == key {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}
	var block []string
	for _, l := range lines[start:] {
		if l != "" && !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
			break
		}
		block = append(block, l)
	}
	return strings.Join(block, "\n")
}

// splitItems splits a list block into `- id:`-anchored items (the same discipline as the
// digest parser: anchor on `- id:` so nested block-list values don't mis-split).
func splitItems(block string) []string {
	re := regexp.MustCompile(`^\s*-\s+id:`)
	var items []string
	var cur []string
	for _, l := range strings.Split(block, "\n") {
		if re.MatchString(l) {
			if len(cur) > 0 {
				items = append(items, strings.Join(cur, "\n"))
			}
			cur = []string{l}
			continue
		}
		if len(cur) > 0 {
			cur = append(cur, l)
		}
	}
	if len(cur) > 0 {
		items = append(items, strings.Join(cur, "\n"))
	}
	return items
}

// ── portfolio numbering ──

var numPrefix = regexp.MustCompile(`^(\d+)-`)

// nextPortfolioNumber returns one past the highest NN- prefix in portfolioDir (1 if none).
func nextPortfolioNumber(portfolioDir string) int {
	files, _ := filepath.Glob(filepath.Join(portfolioDir, "*.md"))
	max := 0
	for _, f := range files {
		m := numPrefix.FindStringSubmatch(filepath.Base(f))
		if m == nil {
			continue
		}
		n := 0
		fmt.Sscanf(m[1], "%d", &n)
		if n > max {
			max = n
		}
	}
	return max + 1
}

// ── small helpers ──

func inSet(xs []string, x string) bool {
	for _, v := range xs {
		if strings.TrimSpace(v) == x {
			return true
		}
	}
	return false
}

func orUnset(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(unset)"
	}
	return s
}

// FindBriefs returns campaign-brief files in a Campaigns dir, newest name first.
func FindBriefs(campaignsDir string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(campaignsDir, "*.md"))
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	return files, nil
}
