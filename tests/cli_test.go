package tests

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"kf2/src/cargo"
	"kf2/src/cli"
)

const (
	exitSuccess = 0
	exitFailure = 1
)

// TestCLIOutput проверяет прямые зависимости и отсутствие вывода параметров этапа 1.
func TestCLIOutput(t *testing.T) {
	deps := []cargo.Dependency{
		{Name: "serde_core", Requirement: "^1.0", Kind: "normal"},
		{Name: "derive", Package: "serde_derive", Requirement: "=1.0.0", Optional: true},
		{Name: "cc", Requirement: "^1", Kind: "build", Target: "cfg(unix)"},
		{Name: "test_crate", Requirement: "*", Kind: "dev", Registry: "https://example.org/index"},
	}
	path := remoteConfig(t, indexLine(t, map[string]any{"deps": deps}))
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", path}, &stdout, &stderr)
	want := "Пакет: serde@1.0.0\nПрямые зависимости:\n" +
		"- serde_core ^1.0 [kind=normal optional=false]\n" +
		"- serde_derive =1.0.0 [kind=normal optional=true alias=derive]\n" +
		"- cc ^1 [kind=build optional=false target=\"cfg(unix)\"]\n" +
		"- test_crate * [kind=dev optional=false registry=\"https://example.org/index\"]\n"
	if status != exitSuccess || stdout.String() != want || stderr.String() != "" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
	}
}

// TestCLIErrors проверяет код ошибки, stderr и пустой stdout.
func TestCLIErrors(t *testing.T) {
	path := writeConfig(t, t.TempDir(), replaceField("filter", "&#10;"))
	cases := [][]string{
		{"-config", path}, {"-config", path + ".missing"}, {"-unknown"}, {"extra"}, {"-config"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := cli.Run(args, &stdout, &stderr)
			if status != exitFailure || stdout.String() != "" ||
				!strings.Contains(stderr.String(), "Ошибка:") {
				t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
			}
		})
	}
}

// TestCLIHelp проверяет успешный вывод справки.
func TestCLIHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if status := cli.Run([]string{"-h"}, &stdout, &stderr); status != exitSuccess {
		t.Fatalf("ожидался успешный код справки; получено %d", status)
	}
	if stdout.String() != "" || !strings.Contains(stderr.String(), "-config") {
		t.Fatalf("неверная справка: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

type failingWriter struct{}

// Write имитирует недоступный поток вывода.
func (failingWriter) Write(data []byte) (int, error) {
	return 0, errors.New("output unavailable")
}

// TestCLIOutputFailure проверяет обработку сбоя записи в stdout.
func TestCLIOutputFailure(t *testing.T) {
	path := remoteConfig(t, indexLine(t, nil))
	var stderr bytes.Buffer
	if status := cli.Run([]string{"-config", path}, failingWriter{}, &stderr); status != exitFailure {
		t.Fatalf("ожидалась ошибка вывода; получено %d", status)
	}
	if !strings.Contains(stderr.String(), "Ошибка вывода:") {
		t.Fatalf("неверная ошибка: %q", stderr.String())
	}
}

// TestCLIEmptyDependencies проверяет явный вывод отсутствия зависимостей.
func TestCLIEmptyDependencies(t *testing.T) {
	path := remoteConfig(t, indexLine(t, nil))
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"--config", path}, &stdout, &stderr)
	if status != exitSuccess || stdout.String() != "Пакет: serde@1.0.0\nПрямые зависимости:\n(нет)\n" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
	}
}

// TestCLIFetchFailure проверяет отсутствие частичного вывода при ошибках получения данных.
func TestCLIFetchFailure(t *testing.T) {
	for _, body := range []string{"broken JSON", indexLine(t, nil) + "broken JSON"} {
		path := remoteConfig(t, body)
		var stdout, stderr bytes.Buffer
		status := cli.Run([]string{"-config", path}, &stdout, &stderr)
		if status != exitFailure || stdout.String() != "" || !strings.Contains(stderr.String(), "repository:") {
			t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
		}
	}
}

// TestCLITestMode проверяет явное сообщение об этапе реализации тестового графа.
func TestCLITestMode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", "../examples/test.xml"}, &stdout, &stderr)
	if status != exitFailure || stdout.String() != "" || !strings.Contains(stderr.String(), "test_mode:") {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
	}
}
