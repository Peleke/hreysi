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
		content = string(data)
		if !strings.Contains(content, "## Commits") {
			content = strings.TrimRight(content, "\n") + "\n\n## Commits\n"
		}
		content += block
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
