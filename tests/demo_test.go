package tests

import (
	"bytes"
	"strings"
	"testing"

	"kf2/src/cargo"
	"kf2/src/cli"
)

// TestStage2Demo демонстрирует прямые зависимости через HTTP без доступа к интернету.
func TestStage2Demo(t *testing.T) {
	deps := []cargo.Dependency{
		{Name: "serde_core", Requirement: "^1.0", Kind: "normal"},
		{Name: "derive", Package: "serde_derive", Requirement: "^1.0", Optional: true},
		{Name: "cc", Requirement: "^1", Kind: "build", Target: "cfg(unix)"},
		{Name: "test_crate", Requirement: "*", Kind: "dev"},
	}
	for _, test := range []struct {
		name string
		deps []cargo.Dependency
	}{
		{"all declared dependencies", deps}, {"no dependencies", []cargo.Dependency{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := serveIndex(t, indexLine(t, map[string]any{"deps": test.deps}))
			content := strings.ReplaceAll(replaceField("filter", "serde"), "https://index.crates.io", server.URL)
			path := writeConfig(t, t.TempDir(), content)
			var stdout, stderr bytes.Buffer
			if status := cli.Run([]string{"-config", path}, &stdout, &stderr); status != exitSuccess {
				t.Fatalf("status=%d stderr=%s", status, stderr.String())
			}
			t.Log("\n" + stdout.String())
			if got := strings.Count(stdout.String(), "\n- "); got != len(test.deps) {
				t.Fatal("выведены не все прямые зависимости")
			}
		})
	}
}
