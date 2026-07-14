// Package mirror copies expanded buildlog entries into an Obsidian vault, so a
// single cross-project corpus exists for downstream consumers (a weekly digest, a
// newsletter, a campaign brief) to read.
//
// It is deliberately one-way. hreysi writes the vault and never reads it back as
// truth: entries the user wrote by hand, or that a predecessor tool wrote, are not
// hreysi's to own. The frontmatter marker below is the whole safety story — a file
// without it is somebody else's file, and mirror refuses to touch it.
//
// Mirroring is OPTIONAL. With no vault configured this package is a no-op and
// hreysi behaves exactly as it does without Obsidian. That is the point: the
// capture story must not depend on anyone's note-taking setup.
package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Peleke/hreysi/internal/entry"
)

// SourceMarker identifies a file in the vault as hreysi's to overwrite. Its ABSENCE
// is what protects a hand-written note (or one from the deprecated buildlog-template)
// from being clobbered on a re-mirror.
const SourceMarker = "source: hreysi"

// VaultSubdir is where mirrored entries land inside the vault. Matches the existing
// dated-entry convention (Buildlog/YYYY-MM-DD-<slug>.md).
const VaultSubdir = "Buildlog"

// vaultConfig is a plain one-line file holding the vault path — no TOML, no parser,
// no dependency. hreysi ships with zero dependencies and this is not worth breaking
// that for.
const vaultConfig = "vault"

// Result reports what a mirror run did.
type Result struct {
	VaultDir string
	Mirrored []string // dest paths written
	Skipped  []string // "<file>: <why>" — nothing to mirror (e.g. no narrative yet)
	Refused  []string // "<file>: <why>" — a foreign file we will not overwrite
}

// VaultDir resolves the vault, or "" when none is configured (mirroring off).
//
// Order: HREYSI_VAULT_DIR, then buildlog/.hreysi/vault. Env wins so a scheduled job
// can point at a vault without touching the repo.
func VaultDir(root string) string {
	if v := strings.TrimSpace(os.Getenv("HREYSI_VAULT_DIR")); v != "" {
		return expandHome(v)
	}
	data, err := os.ReadFile(filepath.Join(root, entry.DirName, ".hreysi", vaultConfig))
	if err != nil {
		return ""
	}
	return expandHome(strings.TrimSpace(string(data)))
}

func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

// SetVault records the vault path for this repo.
func SetVault(root, vault string) error {
	dir := filepath.Join(root, entry.DirName, ".hreysi")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, vaultConfig), []byte(strings.TrimSpace(vault)+"\n"), 0o644)
}

// hasNarrative reports whether an entry is worth mirroring.
//
// A bare `## Commits` spine is a mechanical index, not a story: mirroring it would
// pollute the corpus with entries a digest can draw nothing from. Only entries that
// expansion has touched — v2 (YAML frontmatter) or v1 (`## The Journey`) — carry
// anything a downstream reader wants.
func hasNarrative(content string) bool {
	return strings.HasPrefix(content, "---\n") || strings.Contains(content, "## The Journey")
}

// isOurs reports whether a destination file was written by hreysi and may be
// overwritten. Anything else — a hand-written note, a buildlog-template entry — is
// off limits, permanently.
func isOurs(dest string) (exists, ours bool) {
	data, err := os.ReadFile(dest)
	if err != nil {
		return false, false
	}
	return true, strings.Contains(string(data), SourceMarker)
}

// stamp injects hreysi's provenance keys into the entry's frontmatter, creating the
// frontmatter block if the entry predates the v2 schema. Existing keys are preserved
// verbatim — expansion owns the content, mirror only records where it came from.
func stamp(content, repo, date string) string {
	prov := fmt.Sprintf("%s\nrepo: %s\nmirrored_from: %s/%s.md", SourceMarker, repo, entry.DirName, date)

	if !strings.HasPrefix(content, "---\n") {
		// v1 entry (no frontmatter). Give it a minimal block so the corpus is uniform.
		return fmt.Sprintf("---\ndate: %s\n%s\n---\n\n%s", date, prov, content)
	}
	// v2 entry: splice provenance into the existing block, just before its close.
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		// Malformed frontmatter — don't try to be clever, prepend a fresh block.
		return fmt.Sprintf("---\ndate: %s\n%s\n---\n\n%s", date, prov, content)
	}
	head := rest[:end]
	tail := rest[end:]
	return "---\n" + head + "\n" + prov + tail
}

// Run mirrors every expanded entry in the repo into the vault.
//
// Writes are atomic (temp + rename) because the vault is typically iCloud-synced and
// a half-written file is worse than none.
func Run(root string) (Result, error) {
	res := Result{VaultDir: VaultDir(root)}
	if res.VaultDir == "" {
		return res, nil // mirroring not configured — a no-op, by design
	}

	entries, err := filepath.Glob(filepath.Join(root, entry.DirName, "*.md"))
	if err != nil {
		return res, err
	}
	destDir := filepath.Join(res.VaultDir, VaultSubdir)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return res, err
	}

	repo := filepath.Base(root)
	for _, src := range entries {
		date := strings.TrimSuffix(filepath.Base(src), ".md")
		data, err := os.ReadFile(src)
		if err != nil {
			return res, err
		}
		content := string(data)

		if !hasNarrative(content) {
			res.Skipped = append(res.Skipped, fmt.Sprintf("%s: commit spine only — run `expand` first", date))
			continue
		}

		dest := filepath.Join(destDir, fmt.Sprintf("%s-%s.md", date, repo))
		switch exists, ours := isOurs(dest); {
		case exists && !ours:
			res.Refused = append(res.Refused, fmt.Sprintf("%s: not hreysi's file (no `%s`) — left untouched", filepath.Base(dest), SourceMarker))
			continue
		default:
			if err := atomicWrite(dest, stamp(content, repo, date)); err != nil {
				return res, err
			}
			res.Mirrored = append(res.Mirrored, dest)
		}
	}
	return res, nil
}

func atomicWrite(dest, content string) error {
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".hreysi-mirror-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), dest)
}
