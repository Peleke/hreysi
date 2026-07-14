package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func realBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "hreysi")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

// Regression: `hreysi init` reported "up-to-date" for a hook whose binary was
// gone (/tmp/hreysi-build/hreysi), because idempotency keyed on the marker
// comment alone. Doctor told you to run init to repoint the hook; init refused.
// Capture stayed dead and both commands claimed everything was fine.
func TestInit_RepairsStaleHookTarget(t *testing.T) {
	dir := gitInit(t)
	hooks := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	hook := filepath.Join(hooks, "post-commit")
	stale := "#!/bin/sh\n" + hookBody("/tmp/hreysi-build/hreysi")
	if err := os.WriteFile(hook, []byte(stale), 0o755); err != nil {
		t.Fatal(err)
	}

	bin := realBinary(t)
	res, err := Init(dir, bin)
	if err != nil {
		t.Fatal(err)
	}
	if res.HookAction == "up-to-date" {
		t.Fatal("init refused to repair a hook pointing at a nonexistent binary")
	}

	data, err := os.ReadFile(hook)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "/tmp/hreysi-build/hreysi") {
		t.Error("stale target survived the repair")
	}
	if !strings.Contains(string(data), bin) {
		t.Errorf("hook was not repointed at %s", bin)
	}

	rep, err := Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Healthy {
		t.Errorf("doctor still unhealthy after repair: %+v", rep.Results)
	}
}

// init may have appended hreysi to a hook the user already owned. Repairing our
// target must not cost them their own lines.
func TestInit_RepairPreservesForeignHookLines(t *testing.T) {
	dir := gitInit(t)
	hooks := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	hook := filepath.Join(hooks, "post-commit")
	existing := "#!/bin/sh\n" +
		"echo 'user hook line one'\n" +
		"./scripts/notify.sh\n\n" +
		hookBody("/tmp/hreysi-build/hreysi") +
		"echo 'user hook trailer'\n"
	if err := os.WriteFile(hook, []byte(existing), 0o755); err != nil {
		t.Fatal(err)
	}

	bin := realBinary(t)
	if _, err := Init(dir, bin); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(hook)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, keep := range []string{"user hook line one", "./scripts/notify.sh", "user hook trailer"} {
		if !strings.Contains(got, keep) {
			t.Errorf("repair destroyed a user line: %q\n--- hook ---\n%s", keep, got)
		}
	}
	if strings.Contains(got, "/tmp/hreysi-build/hreysi") {
		t.Error("stale target survived the repair")
	}
	if strings.Count(got, hookMarker) != 1 {
		t.Errorf("marker duplicated; want exactly 1, got %d", strings.Count(got, hookMarker))
	}
}

// An already-correct hook must still be a no-op.
func TestInit_UpToDateWhenTargetGood(t *testing.T) {
	dir := gitInit(t)
	bin := realBinary(t)
	if _, err := Init(dir, bin); err != nil {
		t.Fatal(err)
	}
	res, err := Init(dir, bin) // second run
	if err != nil {
		t.Fatal(err)
	}
	if res.HookAction != "up-to-date" {
		t.Errorf("HookAction = %q, want \"up-to-date\" on a healthy re-run", res.HookAction)
	}
}
