package entry

import (
	"os"
	"strings"
	"testing"

	"github.com/Peleke/hreysi/internal/gitx"
)

func TestBlockContainsEverything(t *testing.T) {
	b := Block(gitx.HeadInfo{
		Hash:      "a1b2c3d",
		Subject:   "feat: add capture",
		Timestamp: "2026-07-06T21:51:54-04:00",
		Date:      "2026-07-06",
		Files:     []string{"main.go", "internal/entry/entry.go"},
	})
	for _, want := range []string{
		"a1b2c3d", "feat: add capture", "2026-07-06T21:51:54-04:00", "`main.go`",
	} {
		if !strings.Contains(b, want) {
			t.Errorf("block missing %q:\n%s", want, b)
		}
	}
}

func TestAppendCreatesThenAppendsSameDay(t *testing.T) {
	root := t.TempDir()

	first := gitx.HeadInfo{Hash: "aaa", Subject: "first", Timestamp: "2026-07-06T10:00:00-04:00", Date: "2026-07-06", Files: []string{"a.go"}}
	path, err := Append(root, first)
	if err != nil {
		t.Fatalf("Append first: %v", err)
	}

	got, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(got), "# 2026-07-06") {
		t.Errorf("missing header:\n%s", got)
	}

	second := gitx.HeadInfo{Hash: "bbb", Subject: "second", Timestamp: "2026-07-06T11:00:00-04:00", Date: "2026-07-06", Files: []string{"b.go"}}
	if _, err := Append(root, second); err != nil {
		t.Fatalf("Append second: %v", err)
	}

	got, _ = os.ReadFile(path)
	s := string(got)
	if n := strings.Count(s, "## Commits"); n != 1 {
		t.Errorf("want exactly one Commits section, got %d:\n%s", n, s)
	}
	if !strings.Contains(s, "aaa") || !strings.Contains(s, "bbb") {
		t.Errorf("both commits should be present:\n%s", s)
	}
}

func TestAppendCapsFileList(t *testing.T) {
	root := t.TempDir()
	files := make([]string, 25)
	for i := range files {
		files[i] = "f" + string(rune('a'+i)) + ".go"
	}
	info := gitx.HeadInfo{Hash: "big", Subject: "many files", Date: "2026-07-06", Files: files}
	path, err := Append(root, info)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "...and 5 more") {
		t.Errorf("expected overflow note for 25 files:\n%s", got)
	}
}
