package ambient

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readHooks(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("settings.json not valid JSON: %v\n%s", err, data)
	}
	hooks, _ := m["hooks"].(map[string]any)
	return hooks
}

func TestMergeCreatesAndIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	if err := mergeSettings(path, []string{"SessionEnd"}); err != nil {
		t.Fatal(err)
	}
	if arr, _ := readHooks(t, path)["SessionEnd"].([]any); len(arr) != 1 {
		t.Fatalf("want 1 SessionEnd entry, got %d", len(arr))
	}

	// Second run must not duplicate.
	if err := mergeSettings(path, []string{"SessionEnd"}); err != nil {
		t.Fatal(err)
	}
	if arr, _ := readHooks(t, path)["SessionEnd"].([]any); len(arr) != 1 {
		t.Errorf("merge not idempotent: %d SessionEnd entries", len(arr))
	}
}

func TestMergePreservesExistingHooks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	seed := `{"hooks":{"SessionEnd":[{"matcher":"","hooks":[{"type":"command","command":"echo existing"}]}]}}`
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := mergeSettings(path, []string{"SessionEnd"}); err != nil {
		t.Fatal(err)
	}
	arr, _ := readHooks(t, path)["SessionEnd"].([]any)
	if len(arr) != 2 {
		t.Fatalf("want 2 entries (existing + hreysi), got %d", len(arr))
	}
	data, _ := os.ReadFile(path)
	if !contains(string(data), "echo existing") || !contains(string(data), "hreysi-expand") {
		t.Errorf("expected both hooks present:\n%s", data)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
