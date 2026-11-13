package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"kf2/src/cli"
)

const stage5Repository = "https://index.crates.io"

var stage5Packages = []string{"windows-link", "windows-sys", "windows-targets"}

// TestStage5Demo воспроизводит три настоящих Cargo-графа по сохранённым метаданным без интернета.
func TestStage5Demo(t *testing.T) {
	server := httptest.NewServer(http.FileServer(http.Dir("testdata/stage5/index")))
	defer server.Close()
	for _, name := range stage5Packages {
		t.Run(name, func(t *testing.T) {
			path := stage5Config(t, server.URL, name)
			var stdout, stderr bytes.Buffer
			status := cli.Run([]string{"-config", path}, &stdout, &stderr)
			if status != exitSuccess || stderr.Len() != noOutput {
				t.Fatalf("status=%d stderr=%s", status, stderr.String())
			}
			got := strings.ReplaceAll(stdout.String(), server.URL, stage5Repository)
			want, err := os.ReadFile("testdata/stage5/graphs/" + name + ".dot")
			if err != nil || got != string(want) {
				t.Fatalf("DOT отличается от примера: %v\n%s", err, got)
			}
			t.Log("\n" + got)
		})
	}
}

func stage5Config(t *testing.T, repository, name string) string {
	t.Helper()
	content := strings.ReplaceAll(validXML, "serde", name)
	content = strings.ReplaceAll(content, stage5Repository, repository)
	return writeConfig(t, t.TempDir(), content)
}
