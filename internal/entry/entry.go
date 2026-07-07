// Package entry renders commits into the dated markdown journal.
//
// The journal directory is the entire product surface and the decoupling
// boundary: hreysi only ever writes it, and any consumer (a narrative-expansion
// skill, a linwheel drafter, a future learning loop) only ever reads it.
package entry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Peleke/hreysi/internal/gitx"
)

// maxFiles caps the file list per commit block to keep entries readable.
const maxFiles = 20

// DirName is the journal directory, relative to the repo root. It is the
// protocol contract every downstream consumer reads against.
const DirName = "buildlog"

// Block renders one commit as a markdown block.
func Block(info gitx.HeadInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\n### `%s` — %s\n", info.Hash, info.Subject)
	if info.Timestamp != "" {
		fmt.Fprintf(&b, "_%s_\n", info.Timestamp)
	}
	if len(info.Files) > 0 {
		b.WriteString("\nFiles:\n")
		shown := info.Files
		if len(shown) > maxFiles {
			shown = shown[:maxFiles]
		}
		for _, f := range shown {
			fmt.Fprintf(&b, "- `%s`\n", f)
		}
		if len(info.Files) > maxFiles {
			fmt.Fprintf(&b, "- ...and %d more\n", len(info.Files)-maxFiles)
		}
	}
	b.WriteString("\n")
	return b.String()
}

// Path is the entry file for a given repo root and YYYY-MM-DD date.
func Path(root, date string) string {
	return filepath.Join(root, DirName, date+".md")
}

// Append writes the commit block into the day's entry, creating the file (and a
// "## Commits" section) if needed. It returns the path written.
func Append(root string, info gitx.HeadInfo) (string, error) {
	if err := os.MkdirAll(filepath.Join(root, DirName), 0o755); err != nil {
		return "", err
	}
	path := Path(root, info.Date)
	block := Block(info)

	var content string
	switch data, err := os.ReadFile(path); {
	case err == nil:
		content = insertCommit(string(data), block)
	case os.IsNotExist(err):
		content = fmt.Sprintf("# %s\n\n## Commits\n%s", info.Date, block)
	default:
		return "", err
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// removeCommitBlock deletes the block for the given short hash from the
// ## Commits section, if present. Used to replace a commit's block on amend
// rather than leave a stale duplicate.
func removeCommitBlock(content, short string) string {
	marker := "### `" + short + "`"
	i := strings.Index(content, marker)
	if i < 0 {
		return content
	}
	rest := content[i+len(marker):]
	end := len(rest)
	if j := strings.Index(rest, "\n### "); j >= 0 && j < end {
		end = j
	}
	if k := strings.Index(rest, "\n## "); k >= 0 && k < end {
		end = k
	}
	before := strings.TrimRight(content[:i], "\n")
	after := strings.TrimLeft(content[i+len(marker)+end:], "\n")
	switch {
	case before == "":
		return after
	case after == "":
		return before + "\n"
	default:
		return before + "\n\n" + after
	}
}

// AppendReplacing removes the block for replaceShort (if present) and then
// inserts the new commit block — used when a commit was amended, so the entry
// shows the amended commit once instead of both.
func AppendReplacing(root string, info gitx.HeadInfo, replaceShort string) (string, error) {
	if err := os.MkdirAll(filepath.Join(root, DirName), 0o755); err != nil {
		return "", err
	}
	path := Path(root, info.Date)
	block := Block(info)

	switch data, err := os.ReadFile(path); {
	case err == nil:
		content := insertCommit(removeCommitBlock(string(data), replaceShort), block)
		if werr := os.WriteFile(path, []byte(content), 0o644); werr != nil {
			return "", werr
		}
	case os.IsNotExist(err):
		content := fmt.Sprintf("# %s\n\n## Commits\n%s", info.Date, block)
		if werr := os.WriteFile(path, []byte(content), 0o644); werr != nil {
			return "", werr
		}
	default:
		return "", err
	}
	return path, nil
}

// insertCommit places a commit block at the end of the "## Commits" section,
// before any narrative section that expansion may have added (## The Journey,
// ## Improvements, ...). Capture (mechanical) and expansion (narrative) own
// different sections of the same file, so a later commit must not append past
// the narrative.
func insertCommit(content, block string) string {
	const marker = "## Commits"
	i := strings.Index(content, marker)
	if i < 0 {
		// No Commits section yet — start one at the end.
		return strings.TrimRight(content, "\n") + "\n\n## Commits\n" + block
	}
	// Look for the next top-level section after the Commits heading.
	rest := content[i+len(marker):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		at := i + len(marker) + j
		head := strings.TrimRight(content[:at], "\n")
		tail := strings.TrimLeft(content[at:], "\n")
		return head + "\n" + block + tail
	}
	// Commits is the last section — append at the end of the file.
	return strings.TrimRight(content, "\n") + "\n" + block
}
