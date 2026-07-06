// Package gitx wraps the handful of git reads hreysi needs.
//
// It shells out to the git binary rather than linking a git library: the work
// is a few reads per commit, the git CLI is always present in a repo we are
// hooking, and staying dependency-free keeps the shipped binary a single
// static file.
package gitx

import (
	"os/exec"
	"strings"
)

// HeadInfo is everything hreysi records about a single commit.
type HeadInfo struct {
	Hash      string   // short hash
	Subject   string   // first line of the commit message
	Timestamp string   // committer date, ISO-8601 (e.g. 2026-07-06T21:51:54-04:00)
	Date      string   // committer date, YYYY-MM-DD (used to bucket entries)
	Files     []string // files touched by the commit
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// RepoRoot returns the top-level directory of the git repo containing dir.
func RepoRoot(dir string) (string, error) {
	return run(dir, "rev-parse", "--show-toplevel")
}

// Head reads metadata about the current HEAD commit.
//
// The timestamp is git's committer date, not wall-clock, so a capture fired
// from a post-commit hook still records the true commit time.
func Head(dir string) (HeadInfo, error) {
	var info HeadInfo
	var err error
	if info.Hash, err = run(dir, "rev-parse", "--short", "HEAD"); err != nil {
		return info, err
	}
	if info.Subject, err = run(dir, "log", "-1", "--format=%s"); err != nil {
		return info, err
	}
	if info.Timestamp, err = run(dir, "log", "-1", "--format=%cI"); err != nil {
		return info, err
	}
	if info.Date, err = run(dir, "log", "-1", "--format=%cd", "--date=format:%Y-%m-%d"); err != nil {
		return info, err
	}

	files, ferr := run(dir, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD")
	if ferr != nil || files == "" {
		// Root commit (or empty diff-tree output): list the whole tree instead.
		files, _ = run(dir, "ls-tree", "--name-only", "-r", "HEAD")
	}
	if files != "" {
		info.Files = strings.Split(files, "\n")
	}
	return info, nil
}
