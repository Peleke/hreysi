// Package skillpack installs the skills that ship with hreysi into a project's
// .claude/skills/ directory. The skills are embedded in the binary, so `hreysi
// init` can drop the buildlog-expansion skill alongside the capture hook —
// capture and expansion arrive together, still decoupled at runtime via the
// buildlog/ directory.
package skillpack

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Install writes the embedded skills into destDir (e.g. a project's
// .claude/skills/ or LifeOS's ~/.claude/Skills/), stripping the embed's root
// directory (skills/ or optskills/). It returns the absolute paths written.
func Install(destDir string, src fs.FS) ([]string, error) {
	var written []string
	err := fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := p
		if i := strings.IndexByte(p, '/'); i >= 0 {
			rel = p[i+1:] // drop the embed root (skills/ or optskills/)
		}
		dest := filepath.Join(destDir, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		data, err := fs.ReadFile(src, p)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return err
		}
		written = append(written, dest)
		return nil
	})
	return written, err
}
