package tests

import (
	"bytes"
	"strings"
	"testing"

	"kf2/src/cli"
)

// TestStage3Demo демонстрирует различные тестовые графы через полный интерфейс CLI.
func TestStage3Demo(t *testing.T) {
	for _, test := range graphCases() {
		t.Run(test.name, func(t *testing.T) {
			path := testConfig(t, test.body, "A", test.filter)
			var stdout, stderr bytes.Buffer
			status := cli.Run([]string{"-config", path}, &stdout, &stderr)
			if status != exitSuccess || stderr.Len() != noOutput {
				t.Fatalf("status=%d stderr=%s", status, stderr.String())
			}
			t.Logf("filter=%q\n%s", test.filter, stdout.String())
			if !strings.HasPrefix(stdout.String(), "digraph dependencies {\n") {
				t.Fatal("граф не выведен")
			}
		})
	}
}
