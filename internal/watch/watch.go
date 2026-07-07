// Package watch tails the git reflog so capture fires for commits made ANY way
// — CLI, GUI, amend — because git appends to logs/HEAD on every HEAD move,
// regardless of whether a hook ran. This is the "can't-miss" capture substrate
// that survives core.hooksPath overrides and hook-less clients.
//
// It polls (stat the reflog for growth) rather than using fsnotify, to keep the
// binary dependency-free. Polling is ample for interactive single-repo use.
package watch

import (
	"os"
	"strings"
	"time"

	"github.com/Peleke/hreysi/internal/gitx"
)

// isCommit reports whether a reflog subject describes a commit (including an
// amend), as opposed to a checkout/reset/merge that also moves HEAD.
func isCommit(reflogSubject string) bool {
	return strings.HasPrefix(reflogSubject, "commit")
}

// Run watches root's reflog and calls onCommit whenever a new commit lands. It
// captures the current HEAD once up front, then loops until the process is
// killed. poll is the stat interval.
func Run(root string, poll time.Duration, onCommit func()) error {
	logPath, err := gitx.GitPath(root, "logs/HEAD")
	if err != nil {
		return err
	}

	var lastSize int64
	if fi, statErr := os.Stat(logPath); statErr == nil {
		lastSize = fi.Size()
	}

	onCommit() // capture whatever HEAD is at now (idempotent)

	for {
		time.Sleep(poll)
		fi, statErr := os.Stat(logPath)
		if statErr != nil || fi.Size() == lastSize {
			continue
		}
		lastSize = fi.Size()
		if isCommit(gitx.ReflogSubject(root)) {
			onCommit()
		}
	}
}
