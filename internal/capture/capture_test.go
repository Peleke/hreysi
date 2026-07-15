package capture

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Peleke/hreysi/internal/entry"
)

func gitInit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "t@e.com"},
		{"config", "user.name", "T"},
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

func commit(t *testing.T, dir, name, content, msg string, amend bool) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("add", name)
	if amend {
		run("commit", "--amend", "-m", msg)
	} else {
		run("commit", "-m", msg)
	}
}

func readEntry(t *testing.T, root string) string {
	t.Helper()
	matches, _ := filepath.Glob(filepath.Join(root, entry.DirName, "*.md"))
	if len(matches) != 1 {
		t.Fatalf("want one entry file, got %v", matches)
	}
	data, _ := os.ReadFile(matches[0])
	return string(data)
}

func TestOnceIsIdempotent(t *testing.T) {
	dir := gitInit(t)
	commit(t, dir, "a.go", "a\n", "feat: one", false)

	if out, err := Once(dir); err != nil || out.Action != "captured" {
		t.Fatalf("first Once = %+v, %v; want captured", out, err)
	}
	// Running again for the same HEAD (hook + watcher both fire) must not
	// double-log.
	out, err := Once(dir)
	if err != nil {
		t.Fatal(err)
	}
	if out.Action != "skipped" {
		t.Errorf("second Once action = %q, want skipped", out.Action)
	}
	if n := strings.Count(readEntry(t, dir), "feat: one"); n != 1 {
		t.Errorf("commit logged %d times, want 1", n)
	}
}

func TestOnceReplacesOnAmend(t *testing.T) {
	dir := gitInit(t)
	commit(t, dir, "a.go", "a\n", "feat: original", false)
	if _, err := Once(dir); err != nil {
		t.Fatal(err)
	}

	// Amend the message; capture must replace, not duplicate.
	commit(t, dir, "a.go", "a\nb\n", "feat: amended", true)
	out, err := Once(dir)
	if err != nil {
		t.Fatal(err)
	}
	if out.Action != "amended" {
		t.Errorf("action = %q, want amended", out.Action)
	}

	s := readEntry(t, dir)
	if strings.Contains(s, "feat: original") {
		t.Errorf("stale pre-amend block still present:\n%s", s)
	}
	if n := strings.Count(s, "### `"); n != 1 {
		t.Errorf("want exactly one commit block after amend, got %d:\n%s", n, s)
	}
	if !strings.Contains(s, "feat: amended") {
		t.Errorf("amended commit missing:\n%s", s)
	}
}

// Regression: if an earlier commit was never captured (watcher missed a poll)
// and the latest commit is amended, the amend must remove the block for the
// commit it actually replaced (HEAD@{1}) — not clobber an unrelated earlier
// block because the marker is stale.
func TestOnceAmendDoesNotClobberUncapturedEarlierCommit(t *testing.T) {
	dir := gitInit(t)
	commit(t, dir, "a.go", "a\n", "feat: A", false)
	if _, err := Once(dir); err != nil { // A captured; marker = A
		t.Fatal(err)
	}
	commit(t, dir, "b.go", "b\n", "feat: B", false)         // NOT captured (missed)
	commit(t, dir, "b.go", "bb\n", "feat: B amended", true) // amend of B

	if _, err := Once(dir); err != nil {
		t.Fatal(err)
	}
	s := readEntry(t, dir)
	if !strings.Contains(s, "feat: A") {
		t.Errorf("earlier commit A was clobbered by the amend:\n%s", s)
	}
	if !strings.Contains(s, "feat: B amended") {
		t.Errorf("amended commit missing:\n%s", s)
	}
}
