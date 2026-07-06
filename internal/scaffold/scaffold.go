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
	Warning    string // e.g. a detected hook manager that may override our hook
}

// detectHookManager returns the name of a hook-management tool in use at root,
// or "" if none. These tools can redirect or own the post-commit hook, so we
// surface them rather than silently assume capture is wired.
func detectHookManager(root string) string {
	for _, c := range []struct{ path, name string }{
		{".husky", "husky"},
		{"lefthook.yml", "lefthook"},
		{"lefthook.yaml", "lefthook"},
		{".pre-commit-config.yaml", "pre-commit"},
	} {
		if _, err := os.Stat(filepath.Join(root, c.path)); err == nil {
			return c.name
		}
	}
	return ""
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

	hookDir, err := gitx.HooksDir(root)
	if err != nil {
		return res, err
	}
	if err := os.MkdirAll(hookDir, 0o755); err != nil {
		return res, err
	}
	res.HookPath = filepath.Join(hookDir, "post-commit")
	res.Warning = detectHookManager(root)
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

// CheckResult is one line in a doctor report. Level is "ok" | "warn" | "fail".
type CheckResult struct {
	Name   string
	Level  string
	Detail string
}

// Report is the result of Check: whether ambient capture will actually fire.
type Report struct {
	Root     string
	HooksDir string
	Results  []CheckResult
	Healthy  bool // no "fail" — capture will fire on commit
}

// Check inspects a repo and reports whether hreysi's capture hook is wired into
// the directory git actually runs hooks from. This is the guarantee `hreysi
// doctor` gives: "is capture live?" answered, not assumed.
func Check(dir string) (Report, error) {
	var r Report

	root, err := gitx.RepoRoot(dir)
	if err != nil {
		return r, fmt.Errorf("not a git repository")
	}
	r.Root = root

	hooksDir, err := gitx.HooksDir(root)
	if err != nil {
		return r, err
	}
	r.HooksDir = hooksDir

	add := func(name, level, detail string) {
		r.Results = append(r.Results, CheckResult{name, level, detail})
	}
	lvl := func(ok bool) string {
		if ok {
			return "ok"
		}
		return "fail"
	}

	if hp := gitx.Config(root, "core.hooksPath"); hp != "" {
		add("core.hooksPath override", "ok", "set to "+hp+" — hook installed there, not .git/hooks")
	}

	hookPath := filepath.Join(hooksDir, "post-commit")
	info, statErr := os.Stat(hookPath)
	exists := statErr == nil
	add("post-commit hook present", lvl(exists), hookPath)

	var wired, executable bool
	if exists {
		executable = info.Mode()&0o100 != 0
		if data, e := os.ReadFile(hookPath); e == nil {
			wired = strings.Contains(string(data), hookMarker)
		}
	}
	add("hreysi capture wired", lvl(wired), "hook runs `hreysi capture`")
	add("hook executable", lvl(executable), "")

	if mgr := detectHookManager(root); mgr != "" {
		add("hook manager detected", "warn", mgr+" in use — confirm capture fires with a test commit")
	}

	_, blErr := os.Stat(filepath.Join(root, entry.DirName))
	if blErr == nil {
		add("journal directory", "ok", entry.DirName+"/")
	} else {
		add("journal directory", "warn", entry.DirName+"/ missing (capture will create it)")
	}

	r.Healthy = exists && wired && executable
	return r, nil
}
