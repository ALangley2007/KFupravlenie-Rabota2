package tests

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"kf2/src/cli"
)

const (
	exitSuccess = 0
	exitFailure = 1
)

// TestCLIOutput проверяет формат и полноту вывода параметров.
func TestCLIOutput(t *testing.T) {
	path := writeConfig(t, t.TempDir(), validXML)
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", path}, &stdout, &stderr)
	want := "package=serde\nrepository=https://index.crates.io\ntest_mode=false\nfilter=\n"
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
	path := writeConfig(t, t.TempDir(), validXML)
	var stderr bytes.Buffer
	if status := cli.Run([]string{"-config", path}, failingWriter{}, &stderr); status != exitFailure {
		t.Fatalf("ожидалась ошибка вывода; получено %d", status)
	}
	if !strings.Contains(stderr.String(), "Ошибка вывода:") {
		t.Fatalf("неверная ошибка: %q", stderr.String())
	}
}
