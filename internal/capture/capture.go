// Package capture is the single idempotent "record the current HEAD" operation
// shared by the git hook (`hreysi capture`) and the reflog watcher (`hreysi
// watch`). Idempotency via a marker lets both run at once without double-logging,
// and makes amend safe.
package capture

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Peleke/hreysi/internal/entry"
	"github.com/Peleke/hreysi/internal/gitx"
)

// Outcome reports what Once did.
type Outcome struct {
	Hash   string
	Path   string
	Action string // "captured" | "amended" | "skipped"
}

func markerPath(root string) string {
	return filepath.Join(root, entry.DirName, ".hreysi", "last_captured")
}

func readMarker(root string) string {
	data, _ := os.ReadFile(markerPath(root))
	return strings.TrimSpace(string(data))
}

func writeMarker(root, full string) {
	p := markerPath(root)
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(full+"\n"), 0o644)
}

// Once records the current HEAD into today's entry, idempotently:
//   - a commit already captured is skipped (hook + watcher coexistence),
//   - an amend replaces the prior block instead of appending a duplicate,
//   - otherwise the commit is appended to the ## Commits section.
func Once(root string) (Outcome, error) {
	info, err := gitx.Head(root)
	if err != nil {
		return Outcome{}, err
	}

	last := readMarker(root)
	if info.FullHash != "" && info.FullHash == last {
		return Outcome{Hash: info.Hash, Action: "skipped"}, nil
	}

	var path string
	action := "captured"
	if strings.HasPrefix(gitx.ReflogSubject(root), "commit (amend)") {
		// The block to replace is the commit the amend *actually* replaced —
		// HEAD@{1} — not whatever the marker points at. The marker can be stale
		// (e.g. the watcher missed the intervening commit between polls), and
		// trusting it would delete the wrong, unrelated block.
		replaced := gitx.ShortHash(root, "HEAD@{1}")
		path, err = entry.AppendReplacing(root, info, replaced)
		action = "amended"
	} else {
		path, err = entry.Append(root, info)
	}
	if err != nil {
		return Outcome{}, err
	}

	writeMarker(root, info.FullHash)
	return Outcome{Hash: info.Hash, Path: path, Action: action}, nil
}
