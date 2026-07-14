package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

// Regression: doctor reported "capture is live — every commit will be journaled"
// for a hook pointing at /tmp/hreysi-build/hreysi, a binary that no longer
// existed. The marker was present, the hook was executable, so every check
// passed — while `|| true` swallowed the failure and NOTHING was captured.
// A green doctor must mean capture actually fires.
func TestHookTarget_ParsesInvokedBinary(t *testing.T) {
	cases := map[string]string{
		hookBody("/tmp/hreysi-build/hreysi"):                      "/tmp/hreysi-build/hreysi",
		hookBody("/opt/homebrew/bin/hreysi"):                      "/opt/homebrew/bin/hreysi",
		"# hreysi: capture this commit\nhreysi capture || true\n": "hreysi", // hand-edited, on PATH
	}
	for body, want := range cases {
		got, ok := hookTarget(body)
		if !ok {
			t.Fatalf("failed to parse a target from hook body %q", body)
		}
		if got != want {
			t.Errorf("hookTarget() = %q, want %q", got, want)
		}
	}
}

func TestTargetResolves_MissingBinaryIsNotLive(t *testing.T) {
	// The exact production failure: an absolute path that is simply gone.
	if targetResolves("/tmp/hreysi-build/hreysi") {
		t.Error("a nonexistent binary must NOT resolve — this is the false green")
	}

	// A real, executable file resolves.
	dir := t.TempDir()
	bin := filepath.Join(dir, "hreysi")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !targetResolves(bin) {
		t.Error("an existing executable must resolve")
	}

	// Present but NOT executable — the hook would still fail to run it.
	noexec := filepath.Join(dir, "hreysi-noexec")
	if err := os.WriteFile(noexec, []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if targetResolves(noexec) {
		t.Error("a non-executable file must NOT resolve")
	}
}

func TestDoctor_FailsWhenHookBinaryMissing(t *testing.T) {
	root := gitInit(t)
	hooks := filepath.Join(root, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	// Simulate the real broken repo: marker present, executable, binary gone.
	hook := filepath.Join(hooks, "post-commit")
	body := "#!/bin/sh\n" + hookBody("/tmp/hreysi-build/hreysi")
	if err := os.WriteFile(hook, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}

	r, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, c := range r.Results {
		if c.Name == "hook target resolves" {
			found = true
			if c.Level != "fail" {
				t.Errorf("hook target check = %q, want \"fail\" (binary does not exist)", c.Level)
			}
		}
	}
	if !found {
		t.Fatal("doctor never checked whether the hook's binary resolves — the original bug")
	}
}
