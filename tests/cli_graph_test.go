package tests

import (
	"bytes"
	"strings"
	"testing"

	"kf2/src/cli"
	"kf2/src/config"
)

// TestCLITestGraphErrors проверяет ошибки формата и отсутствующего корня без частичного вывода.
func TestCLITestGraphErrors(t *testing.T) {
	for _, body := range []string{"A: B", "A: b", "B:"} {
		path := testConfig(t, body, "A", "")
		var stdout, stderr bytes.Buffer
		status := cli.Run([]string{"-config", path}, &stdout, &stderr)
		if status != exitFailure || stdout.Len() != noOutput || !strings.Contains(stderr.String(), "repository:") {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
		}
	}
}

// TestCLIFilteredRoot проверяет объяснение пустого графа и отсутствие запроса корневого пакета.
func TestCLIFilteredRoot(t *testing.T) {
	path := remoteConfig(t, "invalid JSON must not be requested")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	content := strings.ReplaceAll(replaceField("filter", "serde"), "https://index.crates.io", cfg.Repository)
	configPath := writeConfig(t, t.TempDir(), content)
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", configPath}, &stdout, &stderr)
	if status != exitSuccess || stderr.Len() != noOutput ||
		!strings.Contains(stdout.String(), "корневой пакет исключён фильтром") {
		t.Fatalf("status=%d stdout=%s stderr=%s", status, stdout.String(), stderr.String())
	}
}
