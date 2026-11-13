package tests

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"strings"
	"testing"

	"kf2/src/cargo"
	"kf2/src/cli"
	"kf2/src/graph"
	"kf2/src/repository"
)

func serveRepository(t *testing.T, packages map[string][]cargo.Dependency) *httptest.Server {
	t.Helper()
	lines := make(map[string]string)
	for name, dependencies := range packages {
		lines[name] = indexLine(t, map[string]any{"name": name, "deps": dependencies})
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		line, ok := lines[path.Base(request.URL.Path)]
		if !ok {
			http.NotFound(writer, request)
			return
		}
		_, _ = writer.Write([]byte(line))
	}))
	t.Cleanup(server.Close)
	return server
}

func cargoGraphFixture(t *testing.T) *httptest.Server {
	t.Helper()
	return serveRepository(t, map[string][]cargo.Dependency{
		"root": {
			{Name: "skip_alias", Package: "shared", Requirement: "^1"},
			{Name: "build", Requirement: "*", Kind: "build"},
			{Name: "shared", Requirement: "^2", Kind: "dev", Target: "cfg(unix)"},
			{Name: "alias", Package: "skip_pkg", Requirement: "*", Optional: true},
		},
		"shared": {{Name: "Root", Requirement: "*"}, {Name: "leaf", Requirement: "*", Optional: true}},
		"build":  {{Name: "leaf", Requirement: "*"}},
		"leaf":   {},
	})
}

// TestCargoGraph проверяет транзитивность, BFS, псевдонимы, фильтр, цикл и общие зависимости по HTTP.
func TestCargoGraph(t *testing.T) {
	server := cargoGraphFixture(t)
	root := repository.CargoRef(server.URL+"/index/", "ROOT")
	source := &recordingSource{Source: repository.Cargo{}}
	got, err := graph.Build(context.Background(), source, root, "skip")
	want := "root:shared,build\nshared:root,leaf\nbuild:leaf\nleaf:"
	if err != nil || graphText(got) != want {
		t.Fatalf("граф %q, ошибка %v; ожидалось %q", graphText(got), err, want)
	}
	var wantOrder []graph.Ref
	for _, node := range got.Nodes {
		wantOrder = append(wantOrder, node.Ref)
		if node.Version != "1.0.0" {
			t.Fatalf("неверная версия: %+v", node)
		}
	}
	if !reflect.DeepEqual(source.loaded, wantOrder) {
		t.Fatalf("неверный порядок запросов: %+v", source.loaded)
	}
}

// TestCargoGraphRegistries проверяет различение одноимённых пакетов разных sparse-реестров.
func TestCargoGraphRegistries(t *testing.T) {
	other := serveRepository(t, map[string][]cargo.Dependency{"shared": {}})
	server := serveRepository(t, map[string][]cargo.Dependency{
		"root": {
			{Name: "shared", Requirement: "*"},
			{Name: "alias", Package: "shared", Requirement: "*", Registry: "sparse+" + other.URL + "/"},
		},
		"shared": {},
	})
	got, err := graph.Build(context.Background(), repository.Cargo{}, repository.CargoRef(server.URL, "root"), "")
	const expectedNodes = 3
	if err != nil || len(got.Nodes) != expectedNodes {
		t.Fatalf("получено %+v, ошибка %v", got, err)
	}
	first, second := got.Nodes[1], got.Nodes[2]
	if first.Name != second.Name || first.Repository != server.URL || second.Repository != other.URL {
		t.Fatalf("потеряны реестры: %+v", got)
	}
}

// TestCLICargoGraph проверяет вывод транзитивного графа из реального HTTP-источника через CLI.
func TestCLICargoGraph(t *testing.T) {
	server := cargoGraphFixture(t)
	content := strings.ReplaceAll(replaceField("filter", "skip"), "serde", "root")
	content = strings.ReplaceAll(content, "https://index.crates.io", server.URL)
	configPath := writeConfig(t, t.TempDir(), content)
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", configPath}, &stdout, &stderr)
	if status != exitSuccess || stderr.Len() != noOutput {
		t.Fatalf("status=%d stderr=%s", status, stderr.String())
	}
	for _, name := range []string{"root", "shared", "build", "leaf"} {
		if !strings.Contains(stdout.String(), name+" ["+server.URL+"]@1.0.0 -> ") {
			t.Fatalf("пакет %s отсутствует: %s", name, stdout.String())
		}
	}
	if strings.Contains(stdout.String(), "skip") {
		t.Fatalf("фильтр не применён: %s", stdout.String())
	}
}

// TestCLITransitiveFailure проверяет пустой stdout при ошибке в глубине HTTP-графа.
func TestCLITransitiveFailure(t *testing.T) {
	server := serveRepository(t, map[string][]cargo.Dependency{
		"serde": {{Name: "missing", Requirement: "*"}},
	})
	content := strings.ReplaceAll(validXML, "https://index.crates.io", server.URL)
	configPath := writeConfig(t, t.TempDir(), content)
	var stdout, stderr bytes.Buffer
	status := cli.Run([]string{"-config", configPath}, &stdout, &stderr)
	if status != exitFailure || stdout.Len() != noOutput || !strings.Contains(stderr.String(), "missing") {
		t.Fatalf("status=%d stdout=%s stderr=%s", status, stdout.String(), stderr.String())
	}
}
