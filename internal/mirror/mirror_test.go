package mirror

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Peleke/hreysi/internal/entry"
)

func repoWithEntry(t *testing.T, date, content string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, entry.DirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, date+".md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

const v2Entry = `---
title: "The config that was never read"
date: 2026-07-14
project: hreysi
threads:
  - id: dead-mcp-config
    thesis: "settings.json mcpServers was never read"
---

# The config that was never read

## The Journey

It looked like duplication. It was worse.
`

const spineOnly = `# 2026-07-14

## Commits

### ` + "`abc1234`" + ` — feat: thing
`

// No vault configured => hreysi behaves exactly as it does without Obsidian.
// Mirroring must never become a dependency of capture.
func TestRun_NoVaultConfiguredIsNoOp(t *testing.T) {
	root := repoWithEntry(t, "2026-07-14", v2Entry)
	t.Setenv("HREYSI_VAULT_DIR", "")

	res, err := Run(root)
	if err != nil {
		t.Fatalf("mirror must not error when unconfigured: %v", err)
	}
	if res.VaultDir != "" || len(res.Mirrored) != 0 {
		t.Errorf("expected a no-op, got %+v", res)
	}
}

func TestRun_MirrorsExpandedEntryWithProvenance(t *testing.T) {
	root := repoWithEntry(t, "2026-07-14", v2Entry)
	vault := t.TempDir()
	t.Setenv("HREYSI_VAULT_DIR", vault)

	res, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Mirrored) != 1 {
		t.Fatalf("expected 1 mirrored entry, got %+v", res)
	}

	// Named <date>-<repo>.md so a week of cross-project work reads as one corpus.
	dest := filepath.Join(vault, VaultSubdir, "2026-07-14-"+filepath.Base(root)+".md")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("mirrored file not at expected path: %v", err)
	}
	got := string(data)

	if !strings.Contains(got, SourceMarker) {
		t.Error("provenance marker missing — a re-mirror could not tell this file was ours")
	}
	// Expansion owns the content; mirror only records provenance.
	if !strings.Contains(got, `title: "The config that was never read"`) {
		t.Error("existing frontmatter was not preserved")
	}
	if !strings.Contains(got, "id: dead-mcp-config") {
		t.Error("threads[] seam was damaged in transit")
	}
	if !strings.Contains(got, "It looked like duplication. It was worse.") {
		t.Error("narrative body was damaged")
	}
	if strings.Count(got, "---\n") < 2 {
		t.Errorf("frontmatter block is malformed:\n%s", got)
	}
}

// A commit spine has no story. Mirroring it would pollute the corpus with entries a
// digest can draw nothing from.
func TestRun_SkipsUnexpandedSpine(t *testing.T) {
	root := repoWithEntry(t, "2026-07-14", spineOnly)
	vault := t.TempDir()
	t.Setenv("HREYSI_VAULT_DIR", vault)

	res, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Mirrored) != 0 {
		t.Error("mirrored a bare commit spine")
	}
	if len(res.Skipped) != 1 {
		t.Errorf("expected the spine to be reported as skipped, got %+v", res)
	}
}

// THE important one. The vault holds 34 entries written by a predecessor tool, plus
// whatever the human wrote by hand. None of them carry our marker, and none of them
// are ours to overwrite — ever.
func TestRun_RefusesToClobberForeignFile(t *testing.T) {
	root := repoWithEntry(t, "2026-07-14", v2Entry)
	vault := t.TempDir()
	t.Setenv("HREYSI_VAULT_DIR", vault)

	destDir := filepath.Join(vault, VaultSubdir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(destDir, "2026-07-14-"+filepath.Base(root)+".md")
	precious := "---\ntitle: \"Hand-written, or from buildlog-template\"\n---\n\n## The gap we closed\n\nIrreplaceable.\n"
	if err := os.WriteFile(dest, []byte(precious), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != precious {
		t.Fatalf("CLOBBERED a foreign file. This is the one thing mirror must never do.\ngot:\n%s", data)
	}
	if len(res.Refused) != 1 {
		t.Errorf("refusal was not reported to the user: %+v", res)
	}
	if len(res.Mirrored) != 0 {
		t.Error("counted a refused file as mirrored")
	}
}

// A foreign note that merely MENTIONS the marker string — in prose, or quoted inside
// a frontmatter value — is not ours. The guard must key on the marker as a real
// top-level frontmatter key, not on its appearance anywhere in the bytes. (Found by
// dogfooding: an expanded entry whose own narrative discusses `source: hreysi` made
// the string appear three times; strings.Contains over the whole file would classify
// a note *about* mirror as mirror's to overwrite.)
func TestRun_MarkerMentionInProseIsNotOwnership(t *testing.T) {
	root := repoWithEntry(t, "2026-07-14", v2Entry)
	vault := t.TempDir()
	t.Setenv("HREYSI_VAULT_DIR", vault)

	destDir := filepath.Join(vault, VaultSubdir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(destDir, "2026-07-14-"+filepath.Base(root)+".md")

	// A hand-written note ABOUT how mirror works. Contains the marker string twice —
	// once in prose, once quoted inside a frontmatter value — but never as a real key.
	foreign := "---\n" +
		"title: \"How the mirror guard works\"\n" +
		"note: \"it writes source: hreysi into frontmatter\"\n" +
		"---\n\n" +
		"The provenance marker is `source: hreysi`. That is how it knows.\n"
	if err := os.WriteFile(dest, []byte(foreign), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := Run(root)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(dest)
	if string(data) != foreign {
		t.Fatalf("CLOBBERED a note that only MENTIONS the marker.\ngot:\n%s", data)
	}
	if len(res.Refused) != 1 {
		t.Errorf("expected the note to be refused, got %+v", res)
	}
}

// Re-expansion must refresh a file we previously wrote. Idempotent, not append-only.
func TestRun_OverwritesOwnFileAndIsIdempotent(t *testing.T) {
	root := repoWithEntry(t, "2026-07-14", v2Entry)
	vault := t.TempDir()
	t.Setenv("HREYSI_VAULT_DIR", vault)

	if _, err := Run(root); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(vault, VaultSubdir, "2026-07-14-"+filepath.Base(root)+".md")
	first, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}

	res, err := Run(root) // second run, unchanged source
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refused) != 0 {
		t.Errorf("mirror refused its OWN file on re-run: %+v", res.Refused)
	}
	second, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Error("re-mirroring an unchanged entry changed the output (provenance is being re-stamped)")
	}
	if strings.Count(string(second), SourceMarker) != 1 {
		t.Errorf("provenance marker duplicated on re-run: %d occurrences", strings.Count(string(second), SourceMarker))
	}
}

// A v1 entry (expanded, but predating the frontmatter schema) still belongs in the
// corpus — it has narrative. Give it a frontmatter block so the corpus is uniform.
func TestRun_AddsFrontmatterToLegacyExpandedEntry(t *testing.T) {
	v1 := "# 2026-07-14\n\n## Commits\n\n### `abc` — x\n\n## The Journey\n\nWhat happened.\n"
	root := repoWithEntry(t, "2026-07-14", v1)
	vault := t.TempDir()
	t.Setenv("HREYSI_VAULT_DIR", vault)

	if _, err := Run(root); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(vault, VaultSubdir, "2026-07-14-"+filepath.Base(root)+".md")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.HasPrefix(got, "---\n") {
		t.Error("no frontmatter added to a v1 entry")
	}
	if !strings.Contains(got, SourceMarker) || !strings.Contains(got, "date: 2026-07-14") {
		t.Errorf("provenance incomplete:\n%s", got)
	}
	if !strings.Contains(got, "What happened.") {
		t.Error("narrative lost")
	}
}

func TestVaultDir_EnvBeatsConfigFile(t *testing.T) {
	root := repoWithEntry(t, "2026-07-14", v2Entry)
	if err := SetVault(root, "/from/config"); err != nil {
		t.Fatal(err)
	}
	if got := VaultDir(root); got != "/from/config" {
		t.Errorf("config file not read: got %q", got)
	}
	// Env wins, so a scheduled job can target a vault without touching the repo.
	t.Setenv("HREYSI_VAULT_DIR", "/from/env")
	if got := VaultDir(root); got != "/from/env" {
		t.Errorf("env did not take precedence: got %q", got)
	}
}
