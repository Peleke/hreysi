package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitInit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	return dir
}

func TestInitCreatesExecutableHookAndIsIdempotent(t *testing.T) {
	dir := gitInit(t)

	// Must be a binary that really exists: init now repoints a hook whose target
	// does not resolve (that is the whole point of the repair path), so a fixture
	// using a hardcoded, absent "/usr/local/bin/hreysi" would be perpetually
	// "repaired" rather than idempotent — and would be lying about the invariant.
	bin := realBinary(t)

	res, err := Init(dir, bin)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if res.HookAction != "created" {
		t.Errorf("action = %q, want created", res.HookAction)
	}

	data, err := os.ReadFile(res.HookPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), bin) || !strings.Contains(string(data), "capture") {
		t.Errorf("hook does not call the capture binary:\n%s", data)
	}
	if fi, _ := os.Stat(res.HookPath); fi.Mode()&0o100 == 0 {
		t.Error("hook is not executable")
	}
	if _, err := os.Stat(res.EntryDir); err != nil {
		t.Errorf("journal dir missing: %v", err)
	}

	res2, err := Init(dir, bin)
	if err != nil {
		t.Fatalf("second Init: %v", err)
	}
	if res2.HookAction != "up-to-date" {
		t.Errorf("second action = %q, want up-to-date", res2.HookAction)
	}
}

func TestInitPreservesExistingHook(t *testing.T) {
	dir := gitInit(t)
	hookPath := filepath.Join(dir, ".git", "hooks", "post-commit")
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\necho existing-hook\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	res, err := Init(dir, "hreysi")
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if res.HookAction != "appended" {
		t.Errorf("action = %q, want appended", res.HookAction)
	}

	data, _ := os.ReadFile(hookPath)
	if !strings.Contains(string(data), "echo existing-hook") {
		t.Errorf("clobbered the user's hook:\n%s", data)
	}
	if !strings.Contains(string(data), "capture") {
		t.Errorf("did not add capture:\n%s", data)
	}
}

func TestInitRejectsNonRepo(t *testing.T) {
	dir := t.TempDir() // no git init
	if _, err := Init(dir, "hreysi"); err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestCheckHealthyAfterInit(t *testing.T) {
	dir := gitInit(t)
	// Point the hook at a binary that actually exists. The previous fixture used a
	// hardcoded "/usr/local/bin/hreysi" that is absent on most machines — it only
	// passed because Check never verified the hook's target resolved, which was the
	// bug (doctor printed "capture is live" over a dead hook). A healthy repo must
	// mean a hook that can really run.
	bin := filepath.Join(t.TempDir(), "hreysi")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Init(dir, bin); err != nil {
		t.Fatal(err)
	}
	rep, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Healthy {
		t.Errorf("expected healthy after init, got: %+v", rep.Results)
	}
}

func TestCheckFailsWithoutInit(t *testing.T) {
	dir := gitInit(t)
	rep, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Healthy {
		t.Error("expected unhealthy before init (no hook)")
	}
}

// The core.hooksPath override was the empirically-proven capture gap. After
// init, Check must report against the overridden dir and still pass.
func TestCheckHonorsHooksPathOverride(t *testing.T) {
	dir := gitInit(t)
	custom := filepath.Join(dir, "team-hooks")
	cmd := exec.Command("git", "config", "core.hooksPath", custom)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config: %v\n%s", err, out)
	}

	res, err := Init(dir, "hreysi")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(res.HookPath) != filepath.Clean(custom) {
		t.Errorf("hook installed at %q, want under %q", res.HookPath, custom)
	}
	rep, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Healthy {
		t.Errorf("expected healthy with hooksPath override honored, got: %+v", rep.Results)
	}
}
