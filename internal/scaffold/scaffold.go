// Package scaffold wires hreysi into a repo: it creates the journal directory
// and installs a non-blocking post-commit hook. There is no pre-commit wall and
// no enforcement — capture is a side-effect of committing, never a gate.
package scaffold

import (
	"fmt"
	"os"
	"os/exec"
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

// hookTarget extracts the binary path a hreysi post-commit hook actually invokes.
//
// The marker being present only proves someone once ran `hreysi init` here — it
// says nothing about whether the binary that hook points at still exists. A hook
// written by a `go build` into a temp dir (or by a hreysi that has since been
// uninstalled, moved, or upgraded out from under the path) keeps its marker and
// keeps silently failing: `|| true` swallows the error so the commit succeeds and
// nothing is ever journaled. Doctor exists to catch exactly that, so it must
// resolve the target, not grep for a comment.
func hookTarget(data string) (string, bool) {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, "capture") || strings.HasPrefix(line, "#") {
			continue
		}
		// hookBody writes the path with %q, so the common case is a quoted path.
		if i := strings.IndexByte(line, '"'); i >= 0 {
			if j := strings.IndexByte(line[i+1:], '"'); j >= 0 {
				return line[i+1 : i+1+j], true
			}
		}
		// Hand-edited hook: first field, e.g. `hreysi capture || true`.
		if f := strings.Fields(line); len(f) > 0 {
			return f[0], true
		}
	}
	return "", false
}

// replaceHookLines swaps hreysi's own two lines (the marker and the invocation
// that follows it) for a freshly-rendered body, leaving every other line in the
// file untouched. init may have appended hreysi to a hook the user already owned;
// repairing our target must never cost them their lines.
func replaceHookLines(existing, body string) string {
	lines := strings.Split(strings.TrimRight(existing, "\n"), "\n")
	out := make([]string, 0, len(lines)+2)
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != hookMarker {
			out = append(out, lines[i])
			continue
		}
		// Replace the marker plus the invocation line directly beneath it.
		out = append(out, strings.Split(strings.TrimRight(body, "\n"), "\n")...)
		if i+1 < len(lines) && strings.Contains(lines[i+1], "capture") {
			i++ // consume the stale invocation
		}
	}
	return strings.Join(out, "\n") + "\n"
}

// targetResolves reports whether the hook's binary is actually runnable — either
// as a path on disk or as a name found on PATH.
func targetResolves(target string) bool {
	if strings.ContainsRune(target, filepath.Separator) {
		info, err := os.Stat(target)
		return err == nil && !info.IsDir() && info.Mode()&0o111 != 0
	}
	_, err := exec.LookPath(target)
	return err == nil
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
			// The marker alone does not mean the hook works. A hook installed by a
			// `go build` into a temp dir, or by a hreysi since moved/upgraded, keeps
			// its marker while pointing at a binary that no longer exists — and
			// `|| true` hides the failure, so capture dies silently. Treating that as
			// "up-to-date" made init unable to repair the very thing doctor tells you
			// to run init for. Repoint whenever the recorded target isn't what we'd
			// write now, or no longer resolves.
			target, found := hookTarget(string(existing))
			if found && target == exePath && targetResolves(target) {
				res.HookAction = "up-to-date"
				break
			}
			repaired := replaceHookLines(string(existing), body)
			if err := os.WriteFile(res.HookPath, []byte(repaired), 0o755); err != nil {
				return res, err
			}
			if err := os.Chmod(res.HookPath, 0o755); err != nil {
				return res, err
			}
			res.HookAction = "repaired"
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
	var hookContent string
	if exists {
		executable = info.Mode()&0o100 != 0
		if data, e := os.ReadFile(hookPath); e == nil {
			hookContent = string(data)
			wired = strings.Contains(hookContent, hookMarker)
		}
	}
	add("hreysi capture wired", lvl(wired), "hook runs `hreysi capture`")
	add("hook executable", lvl(executable), "")

	// The check that actually answers "will capture fire?". A present, executable,
	// correctly-marked hook still journals nothing if the binary it calls is gone.
	if wired {
		target, found := hookTarget(hookContent)
		switch {
		case !found:
			add("hook target resolves", "fail", "could not parse the binary the hook invokes")
		case targetResolves(target):
			add("hook target resolves", "ok", target)
		default:
			add("hook target resolves", "fail",
				target+" — binary not found. The hook fails silently (`|| true`) and NOTHING is captured. Fix: rerun `hreysi init` to repoint it.")
		}
	}

	if mgr := detectHookManager(root); mgr != "" {
		add("hook manager detected", "warn", mgr+" in use — confirm capture fires with a test commit")
	}

	_, blErr := os.Stat(filepath.Join(root, entry.DirName))
	if blErr == nil {
		add("journal directory", "ok", entry.DirName+"/")
	} else {
		add("journal directory", "warn", entry.DirName+"/ missing (capture will create it)")
	}

	// Derive the verdict from the checks themselves rather than restating a subset
	// of them. The old form (`exists && wired && executable`) silently omitted the
	// target-resolves check, so doctor printed "capture is live" over its own ✗ —
	// the precise false green this command exists to prevent. Any future check that
	// can fail is now automatically load-bearing on the verdict.
	r.Healthy = true
	for _, c := range r.Results {
		if c.Level == "fail" {
			r.Healthy = false
			break
		}
	}
	return r, nil
}
