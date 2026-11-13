package tests

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"kf2/src/cli"
	"kf2/src/config"
)

const (
	exitSuccess = 0
	exitFailure = 1
)

// TestCLIOutput проверяет полный DOT с общими зависимостями и циклом без постороннего текста.
func TestCLIOutput(t *testing.T) {
	path := testConfig(t, "A: B C\nB: D\nC: D\nD: A\n", "A", "")
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", path}, &stdout, &stderr)
	want := "digraph dependencies {\n  rankdir=LR;\n  node [shape=box];\n" +
		"  n0 [label=\"A\"];\n  n1 [label=\"B\"];\n  n2 [label=\"C\"];\n  n3 [label=\"D\"];\n" +
		"  n0 -> n1;\n  n0 -> n2;\n  n1 -> n3;\n  n2 -> n3;\n  n3 -> n0;\n}\n"
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

// TestCLIEmptyDependencies проверяет явное объявление одиночной вершины с версией и реестром.
func TestCLIEmptyDependencies(t *testing.T) {
	path := remoteConfig(t, indexLine(t, nil))
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"--config", path}, &stdout, &stderr)
	want := "digraph dependencies {\n  rankdir=LR;\n  node [shape=box];\n" +
		"  n0 [label=\"serde@1.0.0\\n" + cfg.Repository + "\"];\n}\n"
	if status != exitSuccess || stdout.String() != want || stderr.Len() != noOutput {
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

// TestCLITestMode проверяет пример с путём относительно XML и исключением B.
func TestCLITestMode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", "../examples/test.xml"}, &stdout, &stderr)
	want := "digraph dependencies {\n  rankdir=LR;\n  node [shape=box];\n" +
		"  n0 [label=\"A\"];\n  n1 [label=\"C\"];\n  n0 -> n1;\n}\n"
	if status != exitSuccess || stdout.String() != want || stderr.String() != "" {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
	}
}
