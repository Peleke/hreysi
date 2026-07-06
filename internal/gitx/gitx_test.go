package gitx

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// gitInit makes a real, throwaway git repo. No mocks — we exercise the actual
// git binary against real commits, because that is the only thing that proves
// capture works.
func gitInit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func commit(t *testing.T, dir, name, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", name}, {"commit", "-m", msg}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func TestHeadReadsRealCommit(t *testing.T) {
	dir := gitInit(t)
	commit(t, dir, "hello.txt", "feat: first")

	root, err := RepoRoot(dir)
	if err != nil {
		t.Fatalf("RepoRoot: %v", err)
	}
	info, err := Head(root)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if info.Hash == "" {
		t.Error("empty hash")
	}
	if info.Subject != "feat: first" {
		t.Errorf("subject = %q, want %q", info.Subject, "feat: first")
	}
	if info.Timestamp == "" {
		t.Error("empty timestamp — the whole point is a real commit time")
	}
	if len(info.Date) != len("2026-07-06") {
		t.Errorf("date = %q, want YYYY-MM-DD", info.Date)
	}
	var found bool
	for _, f := range info.Files {
		if f == "hello.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("files = %v, want to contain hello.txt", info.Files)
	}
}
