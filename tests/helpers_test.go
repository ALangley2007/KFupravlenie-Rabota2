package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const filePermissions = 0o600

const validXML = `<config>
<package>serde</package>
<repository>https://index.crates.io</repository>
<test_mode>false</test_mode>
<filter></filter>
</config>`

type invalidCase struct {
	name    string
	content string
	want    string
}

func writeConfig(t *testing.T, directory, content string) string {
	t.Helper()
	path := filepath.Join(directory, "config.xml")
	if err := os.WriteFile(path, []byte(content), filePermissions); err != nil {
		t.Fatal(err)
	}
	return path
}

func replaceField(name, value string) string {
	old := map[string]string{
		"package": "serde", "repository": "https://index.crates.io", "test_mode": "false", "filter": "",
	}
	before := fmt.Sprintf("<%s>%s</%s>", name, old[name], name)
	after := fmt.Sprintf("<%s>%s</%s>", name, value, name)
	return strings.ReplaceAll(validXML, before, after)
}
