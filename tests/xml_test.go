package tests

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"kf2/src/config"
)

// TestRequiredFields проверяет пропуски и повторы каждого параметра.
func TestRequiredFields(t *testing.T) {
	fields := map[string]string{
		"package": "serde", "repository": "https://index.crates.io", "test_mode": "false", "filter": "",
	}
	for name, value := range fields {
		element := fmt.Sprintf("<%s>%s</%s>", name, value, name)
		t.Run(name+" missing", func(t *testing.T) {
			assertConfigError(t, strings.ReplaceAll(validXML, element, ""), name+":")
		})
		t.Run(name+" duplicate", func(t *testing.T) {
			assertConfigError(t, strings.ReplaceAll(validXML, element, element+element), name+":")
		})
	}
}

// TestInvalidXML проверяет синтаксис и структуру XML.
func TestInvalidXML(t *testing.T) {
	tests := []string{
		"", "<config>", "<other/>", validXML + validXML, "text" + validXML, validXML + "text",
		strings.ReplaceAll(validXML, "<config>", `<config xmlns="urn:unexpected">`),
		strings.ReplaceAll(validXML, "<config>", `<config mode="true">`),
		strings.ReplaceAll(validXML, "</config>", "<unknown/></config>"),
		strings.ReplaceAll(validXML, "</config>", "text</config>"),
		replaceField("package", "<nested>serde</nested>"),
		strings.ReplaceAll(validXML, "<package>", `<package name="serde">`),
		strings.ReplaceAll(validXML, "<package>", `<package xmlns="urn:unexpected">`),
		replaceField("filter", "&unknown;"),
	}
	for _, content := range tests {
		t.Run(content, func(t *testing.T) {
			assertConfigError(t, content, "конфигурация XML:")
		})
	}
}

// TestValidXMLFormatting проверяет допустимые варианты оформления.
func TestValidXMLFormatting(t *testing.T) {
	content := `<?xml version="1.0"?><!-- before -->` + validXML + "<!-- after -->\n"
	content = strings.ReplaceAll(content, "<filter></filter>", "<filter/>")
	content = strings.ReplaceAll(content, "<package>serde</package>", "<package> serde </package>")
	if _, err := config.Load(writeConfig(t, t.TempDir(), content)); err != nil {
		t.Fatal(err)
	}
}

// TestMissingConfigFile проверяет сообщение об отсутствующем XML.
func TestMissingConfigFile(t *testing.T) {
	_, err := config.Load(filepath.Join(t.TempDir(), "missing.xml"))
	if err == nil || !strings.Contains(err.Error(), "файл конфигурации:") {
		t.Fatalf("ожидалась ошибка чтения конфигурации; получено %v", err)
	}
}
