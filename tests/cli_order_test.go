package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kf2/src/cargo"
	"kf2/src/cli"
	"kf2/src/config"
)

func orderConfig(t *testing.T, test orderCase) string {
	t.Helper()
	path := testConfig(t, test.body, test.rootName(), test.filter)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return writeConfig(t, filepath.Dir(path), withLoadOrder(string(content), "true"))
}

func withLoadOrder(content, value string) string {
	return strings.ReplaceAll(content, "</config>", "<load_order>"+value+"</load_order></config>")
}

// TestStage4Demo демонстрирует и проверяет порядок загрузки через CLI, включая ошибки циклов.
func TestStage4Demo(t *testing.T) {
	for _, test := range orderCases() {
		t.Run(test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			status := cli.Run([]string{"-config", orderConfig(t, test)}, &stdout, &stderr)
			if test.cycle != "" {
				want := "цикл " + strings.ReplaceAll(test.cycle, " ", " -> ")
				if status != exitFailure || stdout.Len() != noOutput || !strings.Contains(stderr.String(), want) {
					t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
				}
				t.Logf("filter=%q\n%s", test.filter, stderr.String())
				return
			}
			want := "Пакет: " + test.rootName() + "\nПорядок загрузки (зависимости перед пакетом):\n"
			if test.want == "" {
				want += "(пуст: корневой пакет исключён фильтром)\n"
			} else {
				want += strings.ReplaceAll(test.want, " ", "\n") + "\n"
			}
			if status != exitSuccess || stderr.Len() != noOutput || stdout.String() != want {
				t.Fatalf("status=%d stdout=%q stderr=%q; want=%q", status, stdout.String(), stderr.String(), want)
			}
			t.Logf("filter=%q\n%s", test.filter, stdout.String())
		})
	}
}

// TestCLILoadOrderWriteFailure проверяет ошибку записи в новом режиме вывода.
func TestCLILoadOrderWriteFailure(t *testing.T) {
	path := orderConfig(t, orderCase{body: "A:"})
	var stderr bytes.Buffer
	status := cli.Run([]string{"-config", path}, failingWriter{}, &stderr)
	if status != exitFailure || !strings.Contains(stderr.String(), "Ошибка вывода:") {
		t.Fatalf("status=%d stderr=%q", status, stderr.String())
	}
}

// TestCLICargoLoadOrder проверяет реальный HTTP-источник, версии и обозначение реестра в порядке загрузки.
func TestCLICargoLoadOrder(t *testing.T) {
	server := serveRepository(t, map[string][]cargo.Dependency{
		"serde": {{Name: "leaf", Requirement: "*"}}, "leaf": {},
	})
	content := strings.ReplaceAll(withLoadOrder(validXML, "true"), "https://index.crates.io", server.URL)
	path := writeConfig(t, t.TempDir(), content)
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", path}, &stdout, &stderr)
	want := "Пакет: serde [" + server.URL + "]\nПорядок загрузки (зависимости перед пакетом):\n" +
		"leaf [" + server.URL + "]@1.0.0\nserde [" + server.URL + "]@1.0.0\n"
	if status != exitSuccess || stderr.Len() != noOutput || stdout.String() != want {
		t.Fatalf("status=%d stdout=%q stderr=%q", status, stdout.String(), stderr.String())
	}
}

// TestLoadOrderConfig проверяет необязательный XML-параметр, его значения и ошибки структуры.
func TestLoadOrderConfig(t *testing.T) {
	for _, value := range []string{"true", "false", " true "} {
		path := writeConfig(t, t.TempDir(), withLoadOrder(validXML, value))
		cfg, err := config.Load(path)
		if err != nil || cfg.LoadOrder != (strings.TrimSpace(value) == "true") {
			t.Fatalf("load_order=%q: %+v, %v", value, cfg, err)
		}
	}
	for _, value := range []string{"", " ", "TRUE", "1", "yes", "<nested/>", "&#10;"} {
		assertConfigError(t, withLoadOrder(validXML, value), "load_order:")
	}
	for _, element := range []string{
		"<load_order>true</load_order><load_order>false</load_order>",
		`<load_order value="true">true</load_order>`,
		`<load_order xmlns="urn:unexpected">true</load_order>`,
	} {
		assertConfigError(t, strings.ReplaceAll(validXML, "</config>", element+"</config>"), "load_order:")
	}
}
