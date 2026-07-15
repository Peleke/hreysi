package digest

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ParseEntry extracts threads[] from one mirrored entry's YAML frontmatter.
//
// hreysi has zero dependencies and no YAML parser, but it OWNS this schema (expand
// writes it), so a line-based parser keyed on the known structure is safe and honest —
// the alternative is taking on a YAML dep for one block we control end to end. It reads
// only the frontmatter block and only the threads[] list; malformed input yields the
// threads it could parse, never a panic.
func ParseEntry(content, repo, date string) []Thread {
	fm := frontmatter(content)
	if fm == "" {
		return nil
	}
	block := threadsBlock(fm)
	if block == "" {
		return nil
	}
	var out []Thread
	for _, item := range splitThreadItems(block) {
		t := parseThread(item)
		if t.ID == "" {
			continue // an item with no id is not a thread
		}
		t.Repo, t.Date = repo, date
		out = append(out, t)
	}
	return out
}

// frontmatter returns the content between the leading `---` fence and its close.
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

// threadsBlock returns the lines under `threads:` — everything indented beneath it, up
// to the next top-level (column-zero) key.
func threadsBlock(fm string) string {
	lines := strings.Split(fm, "\n")
	start := -1
	for i, l := range lines {
		if l == "threads:" || strings.HasPrefix(l, "threads:") && strings.TrimSpace(l) == "threads:" {
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
			break // a new top-level key ends the threads block
		}
		block = append(block, l)
	}
	return strings.Join(block, "\n")
}

// splitThreadItems splits the block into per-thread chunks. A thread starts at a
// list-item line whose first key is id (`  - id: ...`).
func splitThreadItems(block string) []string {
	var items []string
	var cur []string
	// A thread starts with `- id:`. Anchoring on that (not a bare `- `) is essential:
	// block-list values inside a thread (e.g. `      - "hreysi 1.1.1"` under facts:)
	// also begin with `- ` and would otherwise be mis-split into empty pseudo-threads.
	itemStart := regexp.MustCompile(`^\s*-\s+id:`)
	for _, l := range strings.Split(block, "\n") {
		if itemStart.MatchString(l) {
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

var scalarRe = regexp.MustCompile(`^\s*-?\s*(\w+):\s*(.*)$`)

// parseThread reads one thread item. Handles scalar keys, inline arrays (`tags: [a, b]`),
// and block-list arrays (`facts:` then `  - "x"` lines).
func parseThread(item string) Thread {
	var t Thread
	lines := strings.Split(item, "\n")
	for i := 0; i < len(lines); i++ {
		m := scalarRe.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		key, val := m[1], strings.TrimSpace(m[2])
		switch key {
		case "id":
			t.ID = unquote(val)
		case "thesis":
			t.Thesis = unquote(val)
		case "evidence":
			t.Evidence = unquote(val)
		case "section":
			t.Section = unquote(val)
		case "angle":
			t.Angle = unquote(val)
		case "scale":
			t.Scale = unquote(val)
		case "fit":
			t.Fit, _ = strconv.Atoi(strings.TrimSpace(val))
		case "tags":
			t.Tags = parseArray(val, lines, &i)
		case "facts":
			t.Facts = parseArray(val, lines, &i)
		}
	}
	return t
}

// parseArray reads either an inline array on the same line (`[a, b, c]`) or a block list
// on the following more-indented `- ` lines. When it consumes block lines it advances
// the caller's index past them.
func parseArray(inlineVal string, lines []string, i *int) []string {
	inlineVal = strings.TrimSpace(inlineVal)
	if strings.HasPrefix(inlineVal, "[") {
		inner := strings.TrimSuffix(strings.TrimPrefix(inlineVal, "["), "]")
		var out []string
		for _, p := range strings.Split(inner, ",") {
			if p = unquote(strings.TrimSpace(p)); p != "" {
				out = append(out, p)
			}
		}
		return out
	}
	// Block list: consume following `- ...` lines.
	var out []string
	listItem := regexp.MustCompile(`^\s+-\s+(.*)$`)
	for j := *i + 1; j < len(lines); j++ {
		m := listItem.FindStringSubmatch(lines[j])
		if m == nil {
			// Stop at the next key or a non-list line.
			if strings.TrimSpace(lines[j]) == "" {
				continue
			}
			break
		}
		if v := unquote(strings.TrimSpace(m[1])); v != "" {
			out = append(out, v)
		}
		*i = j
	}
	return out
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		s = s[1 : len(s)-1]
	}
	return strings.ReplaceAll(s, `\"`, `"`)
}

// ── corpus + coverage ──

var mirroredName = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)\.md$`)

// ParseCorpus reads every mirrored entry in buildlogDir whose date falls in
// [since, until] (inclusive, YYYY-MM-DD string compare). Entry filenames are
// YYYY-MM-DD-<repo>.md, written by `hreysi mirror`.
func ParseCorpus(buildlogDir, since, until string) ([]Thread, map[string]bool, error) {
	files, err := filepath.Glob(filepath.Join(buildlogDir, "*.md"))
	if err != nil {
		return nil, nil, err
	}
	var threads []Thread
	repos := map[string]bool{}
	for _, f := range files {
		m := mirroredName.FindStringSubmatch(filepath.Base(f))
		if m == nil {
			continue // not a hreysi-mirrored entry (e.g. a legacy hand-named note)
		}
		date, repo := m[1], m[2]
		if date < since || date > until {
			continue
		}
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, nil, err
		}
		// Only our own entries carry threads[]; legacy files parse to nothing and are
		// skipped, which is the intended clean split (they stay mute to digest).
		ts := ParseEntry(string(data), repo, date)
		if len(ts) > 0 {
			repos[repo] = true
			threads = append(threads, ts...)
		}
	}
	return threads, repos, nil
}

// Coverage names what digest could and could not see, so a thin week cannot masquerade
// as complete.
type Coverage struct {
	Threaded  []string // repos with mirrored, threaded entries in the window
	Uncovered []string // repos with recent local buildlog activity but no threaded entry
}

// DetectCoverage compares repos represented in the corpus against repos under
// projectsDir that have RECENT local buildlog entries — best-effort, mechanical. It can
// only report gaps it can see, and says so; it never implies completeness.
func DetectCoverage(projectsDir, since, until string, threadedRepos map[string]bool) Coverage {
	cov := Coverage{}
	for r := range threadedRepos {
		cov.Threaded = append(cov.Threaded, r)
	}
	sort.Strings(cov.Threaded)

	dirs, _ := filepath.Glob(filepath.Join(projectsDir, "*", "buildlog", "*.md"))
	active := map[string]bool{}
	for _, f := range dirs {
		date := strings.TrimSuffix(filepath.Base(f), ".md")
		if date < since || date > until {
			continue
		}
		// projectsDir/<repo>/buildlog/<date>.md
		repo := filepath.Base(filepath.Dir(filepath.Dir(f)))
		active[repo] = true
	}
	for r := range active {
		if !threadedRepos[r] {
			cov.Uncovered = append(cov.Uncovered, r)
		}
	}
	sort.Strings(cov.Uncovered)
	return cov
}
