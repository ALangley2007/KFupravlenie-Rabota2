package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kf2/src/graph"
)

const directoryPermissions = 0o700

func cargoFixture(t *testing.T, dependencies graph.Graph) string {
	t.Helper()
	directory := t.TempDir()
	for _, node := range dependencies.Nodes {
		name := strings.ToLower(node.Name)
		crate := filepath.Join(directory, name)
		if err := os.MkdirAll(filepath.Join(crate, "src"), directoryPermissions); err != nil {
			t.Fatal(err)
		}
		manifest := fmt.Sprintf("[package]\nname = %q\nversion = \"1.0.0\"\nedition = \"2021\"\n", name)
		manifest += "\n[dependencies]\n"
		for _, dependency := range node.Dependencies {
			dependencyName := strings.ToLower(dependency.Name)
			manifest += fmt.Sprintf("%s = { path = \"../%s\" }\n", dependencyName, dependencyName)
		}
		writeFixtureFile(t, filepath.Join(crate, "Cargo.toml"), manifest)
		writeFixtureFile(t, filepath.Join(crate, "src", "lib.rs"), "pub fn value() -> u8 { 1 }\n")
	}
	return filepath.Join(directory, strings.ToLower(dependencies.Root.Name), "Cargo.toml")
}

func writeFixtureFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), filePermissions); err != nil {
		t.Fatal(err)
	}
}

func cargoArtifacts(t *testing.T, data []byte) []string {
	t.Helper()
	var names []string
	for line := range strings.SplitSeq(strings.TrimSpace(string(data)), "\n") {
		var event struct {
			Reason string `json:"reason"`
			Target struct {
				Name string `json:"name"`
			} `json:"target"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event.Reason == "compiler-artifact" {
			names = append(names, strings.ToUpper(event.Target.Name))
		}
	}
	return names
}

func nodesInCargoOrder(t *testing.T, dependencies graph.Graph, names []string) []graph.Node {
	t.Helper()
	nodes := make(map[string]graph.Node)
	for _, node := range dependencies.Nodes {
		nodes[node.Name] = node
	}
	var order []graph.Node
	for _, name := range names {
		node, ok := nodes[name]
		if !ok {
			t.Fatalf("Cargo вернул неизвестный пакет %s", name)
		}
		order = append(order, node)
	}
	if len(order) != len(dependencies.Nodes) {
		t.Fatalf("Cargo вернул %d пакетов, ожидалось %d", len(order), len(dependencies.Nodes))
	}
	return order
}
