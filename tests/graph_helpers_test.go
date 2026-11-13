package tests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kf2/src/graph"
	"kf2/src/repository"
)

const noOutput = 0

type recordingSource struct {
	graph.Source
	loaded []graph.Ref
}

// Load записывает последовательность загрузки, затем обращается к исходному репозиторию.
func (source *recordingSource) Load(ctx context.Context, ref graph.Ref) (graph.Node, error) {
	source.loaded = append(source.loaded, ref)
	return source.Source.Load(ctx, ref)
}

func repositoryFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "repository.txt")
	if err := os.WriteFile(path, []byte(body), filePermissions); err != nil {
		t.Fatal(err)
	}
	return path
}

func testSource(t *testing.T, body string) *recordingSource {
	t.Helper()
	source, err := repository.LoadTest(repositoryFile(t, body))
	if err != nil {
		t.Fatal(err)
	}
	return &recordingSource{Source: source}
}

func testConfig(t *testing.T, body, root, filter string) string {
	t.Helper()
	path := repositoryFile(t, body)
	content := strings.ReplaceAll(validXML, "serde", root)
	content = strings.ReplaceAll(content, "false", "true")
	content = strings.ReplaceAll(content, "https://index.crates.io", filepath.Base(path))
	content = strings.ReplaceAll(content, "<filter></filter>", "<filter>"+filter+"</filter>")
	return writeConfig(t, filepath.Dir(path), content)
}

func graphText(dependencies graph.Graph) string {
	var lines []string
	for _, node := range dependencies.Nodes {
		var targets []string
		for _, dependency := range node.Dependencies {
			targets = append(targets, dependency.Name)
		}
		lines = append(lines, node.Name+":"+strings.Join(targets, ","))
	}
	return strings.Join(lines, "\n")
}
