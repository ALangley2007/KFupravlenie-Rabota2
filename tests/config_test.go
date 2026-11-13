package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kf2/src/config"
)

// TestLoadRemoteConfig проверяет параметры URL и сохранение фильтра.
func TestLoadRemoteConfig(t *testing.T) {
	path := writeConfig(t, t.TempDir(), replaceField("filter", " keep &amp; skip "))
	got, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Config{
		Package: "serde", Repository: "https://index.crates.io", Filter: " keep & skip ",
	}
	if got != want {
		t.Fatalf("получено %+v; ожидалось %+v", got, want)
	}
}

// TestLoadLocalRepository проверяет абсолютный путь и путь от XML.
func TestLoadLocalRepository(t *testing.T) {
	path := filepath.Join(t.TempDir(), "repository.txt")
	if err := os.WriteFile(path, []byte("A: B\nB:\n"), filePermissions); err != nil {
		t.Fatal(err)
	}
	for _, repository := range []string{path, filepath.Base(path)} {
		t.Run(repository, func(t *testing.T) {
			content := strings.ReplaceAll(validXML, "serde", "A")
			content = strings.ReplaceAll(content, "false", "true")
			content = strings.ReplaceAll(content, "https://index.crates.io", repository)
			cfg, err := config.Load(writeConfig(t, filepath.Dir(path), content))
			if err != nil {
				t.Fatal(err)
			}
			if !cfg.TestMode || cfg.Package != "A" || cfg.Repository != repository {
				t.Fatalf("неверные параметры: %+v", cfg)
			}
		})
	}
}

// TestInvalidParameters проверяет ошибки всех четырёх параметров.
func TestInvalidParameters(t *testing.T) {
	for _, test := range invalidParameterCases() {
		t.Run(test.name, func(t *testing.T) {
			assertConfigError(t, test.content, test.want)
		})
	}
}

func invalidParameterCases() []invalidCase {
	return []invalidCase{
		{"empty package", replaceField("package", ""), "package:"},
		{"blank package", replaceField("package", "  "), "package:"},
		{"invalid package", replaceField("package", "bad/name"), "package:"},
		{"empty repository", replaceField("repository", ""), "repository:"},
		{"local path in remote mode", replaceField("repository", "repo.txt"), "repository:"},
		{"unsupported scheme", replaceField("repository", "ftp://example.org"), "repository:"},
		{"missing host", replaceField("repository", "https:///index"), "repository:"},
		{"invalid URL", replaceField("repository", "https://["), "repository:"},
		{"URL with space", replaceField("repository", "https://example.org/a b"), "repository:"},
		{"repository control character", replaceField("repository", "re&#10;po.txt"), "repository:"},
		{"empty port", replaceField("repository", "https://example.org:"), "repository:"},
		{"zero port", replaceField("repository", "https://example.org:0"), "repository:"},
		{"large port", replaceField("repository", "https://example.org:65536"), "repository:"},
		{"text port", replaceField("repository", "https://example.org:bad"), "repository:"},
		{"empty mode", replaceField("test_mode", ""), "test_mode:"},
		{"invalid mode", replaceField("test_mode", "yes"), "test_mode:"},
		{"numeric mode", replaceField("test_mode", "1"), "test_mode:"},
		{"uppercase mode", replaceField("test_mode", "TRUE"), "test_mode:"},
		{"filter control character", replaceField("filter", "&#10;"), "filter:"},
	}
}

// TestInvalidLocalRepository проверяет ошибки тестового режима.
func TestInvalidLocalRepository(t *testing.T) {
	content := strings.ReplaceAll(replaceField("package", "A"), "false", "true")
	tests := []invalidCase{
		{"test package name", strings.ReplaceAll(content, "<package>A", "<package>serde"), "package:"},
		{"URL in test mode", content, "repository:"},
		{"missing file", strings.ReplaceAll(content, "https://index.crates.io", "missing.txt"), "repository:"},
		{"directory", strings.ReplaceAll(content, "https://index.crates.io", "."), "repository:"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertConfigError(t, test.content, test.want)
		})
	}
}

func assertConfigError(t *testing.T, content, want string) {
	t.Helper()
	_, err := config.Load(writeConfig(t, t.TempDir(), content))
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("ожидалась ошибка с %q; получено %v", want, err)
	}
}
