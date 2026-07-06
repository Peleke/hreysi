// Package scaffold wires hreysi into a repo: it creates the journal directory
// and installs a non-blocking post-commit hook. There is no pre-commit wall and
// no enforcement — capture is a side-effect of committing, never a gate.
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Peleke/hreysi/internal/entry"
	"github.com/Peleke/hreysi/internal/gitx"
)

// hookMarker identifies hreysi's lines inside a post-commit hook so init stays
// idempotent and never clobbers a user's existing hook.
const hookMarker = "# hreysi: capture this commit"

func hookBody(exePath string) string {
	// `|| true` guarantees a hreysi hiccup can never fail a commit.
	return fmt.Sprintf("%s\n%q capture || true\n", hookMarker, exePath)
}

// Result reports what Init did, for a friendly summary.
type Result struct {
	Root       string
	EntryDir   string
	HookPath   string
	HookAction string // "created" | "appended" | "up-to-date"
}

// Init scaffolds the journal directory and installs the post-commit hook.
// exePath is the absolute path to the hreysi binary, baked into the hook so it
// works even when hreysi is not on the hook's PATH.
func Init(dir, exePath string) (Result, error) {
	var res Result

	root, err := gitx.RepoRoot(dir)
	if err != nil {
		return res, fmt.Errorf("not a git repository (run `git init` first)")
	}
	res.Root = root

	res.EntryDir = filepath.Join(root, entry.DirName)
	if err := os.MkdirAll(res.EntryDir, 0o755); err != nil {
		return res, err
	}

	hookDir := filepath.Join(root, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return res, err
	}
	res.HookPath = filepath.Join(hookDir, "post-commit")
	body := hookBody(exePath)

	switch existing, err := os.ReadFile(res.HookPath); {
	case os.IsNotExist(err):
		if err := os.WriteFile(res.HookPath, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
			return res, err
		}
		res.HookAction = "created"
	case err == nil:
		if strings.Contains(string(existing), hookMarker) {
			res.HookAction = "up-to-date"
			break
		}
		merged := strings.TrimRight(string(existing), "\n") + "\n\n" + body
		if err := os.WriteFile(res.HookPath, []byte(merged), 0o755); err != nil {
			return res, err
		}
		if err := os.Chmod(res.HookPath, 0o755); err != nil {
			return res, err
		}
		res.HookAction = "appended"
	default:
		return res, err
	}
	return res, nil
}
