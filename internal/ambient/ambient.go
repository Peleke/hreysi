// Package ambient wires the expansion trigger into Claude Code: it drops the
// hook script and registers it in .claude/settings.json for the chosen lifecycle
// event (SessionEnd by default, Stop behind a flag). The merge preserves any
// hooks already configured — hreysi adds itself, it does not take over.
package ambient

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const scriptRel = ".claude/hooks/hreysi-expand.sh"

// hookCommand is what settings.json runs. $CLAUDE_PROJECT_DIR is expanded by
// Claude Code at hook time.
const hookCommand = `"$CLAUDE_PROJECT_DIR"/.claude/hooks/hreysi-expand.sh`

// Install drops the embedded hook script and registers it for each event
// ("SessionEnd", "Stop"). src is the embedded FS containing hooks/hreysi-expand.sh.
func Install(root string, src fs.FS, events []string) error {
	if err := dropScript(root, src); err != nil {
		return err
	}
	return mergeSettings(filepath.Join(root, ".claude", "settings.json"), events)
}

func dropScript(root string, src fs.FS) error {
	data, err := fs.ReadFile(src, "hooks/hreysi-expand.sh")
	if err != nil {
		return err
	}
	dest := filepath.Join(root, scriptRel)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o755)
}

// mergeSettings adds the hreysi hook to each event's array in settings.json,
// creating the file/keys as needed and skipping events already wired.
func mergeSettings(path string, events []string) error {
	settings := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &settings) // tolerate/overwrite malformed
	} else if !os.IsNotExist(err) {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}

	for _, ev := range events {
		arr, _ := hooks[ev].([]any)
		if hookAlreadyPresent(arr) {
			continue
		}
		arr = append(arr, map[string]any{
			"matcher": "",
			"hooks": []any{
				map[string]any{"type": "command", "command": hookCommand},
			},
		})
		hooks[ev] = arr
	}
	settings["hooks"] = hooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(out, '\n'), 0o644)
}

// hookAlreadyPresent reports whether any entry in the event array already runs
// the hreysi expand hook (so Install is idempotent).
func hookAlreadyPresent(arr []any) bool {
	for _, e := range arr {
		entry, _ := e.(map[string]any)
		inner, _ := entry["hooks"].([]any)
		for _, h := range inner {
			hm, _ := h.(map[string]any)
			if cmd, _ := hm["command"].(string); strings.Contains(cmd, "hreysi-expand") {
				return true
			}
		}
	}
	return false
}
