package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"

	"kf2/src/graph"
	"kf2/src/repository"
)

// TestCargoTreeComparison сравнивает три снимка с настоящим cargo tree, запуская Cargo только для проверки.
func TestCargoTreeComparison(t *testing.T) {
	if os.Getenv("KF2_CARGO_TREE_COMPARE") != "true" {
		t.Skip("для сетевого сравнения с Cargo выполните ./run.sh compare-visual")
	}
	server := httptest.NewServer(http.FileServer(http.Dir("testdata/stage5/index")))
	defer server.Close()
	cache := t.TempDir()
	for _, name := range stage5Packages {
		t.Run(name, func(t *testing.T) {
			root := repository.CargoRef(server.URL, name)
			dependencies, err := graph.Build(context.Background(), repository.Cargo{}, root, "")
			if err != nil {
				t.Fatal(err)
			}
			compareCargoTree(t, dependencies, cache)
		})
	}
}

func compareCargoTree(t *testing.T, dependencies graph.Graph, cache string) {
	t.Helper()
	root := dependencies.Nodes[0]
	manifest := cargoTreeManifest(t, root)
	output := runCargoTree(t, root, manifest, cache, "all", "depth")
	got := parseCargoTree(t, output)
	want := dependencyEdges(dependencies)
	if !slices.Equal(got, want) {
		t.Fatalf("рёбра Cargo: %v; рёбра kf2: %v", got, want)
	}
	names := cargoTreePackages(t, output)
	for _, node := range dependencies.Nodes {
		if _, present := names[node.Name]; !present {
			t.Fatalf("Cargo не показал пакет %s", node.Name)
		}
		if node.Version != names[node.Name] {
			if node.Name != "windows-link" || names[node.Name] != "0.2.1" {
				t.Fatalf("неожиданное расхождение версий: %s: %s / %s", node.Name, node.Version, names[node.Name])
			}
			t.Logf("%s: kf2=%s, Cargo=%s: Cargo учитывает req=^0.2.1", node.Name, node.Version, names[node.Name])
		}
	}
	if len(names) != len(dependencies.Nodes) {
		t.Fatalf("разное число вершин: Cargo=%d; kf2=%d", len(names), len(dependencies.Nodes))
	}
	runCargoTree(t, root, manifest, cache, "all", "indent")
	runCargoTree(t, root, manifest, cache, "", "indent")
}

func dependencyEdges(dependencies graph.Graph) []string {
	edges := []string{}
	for _, node := range dependencies.Nodes {
		for _, dependency := range node.Dependencies {
			edges = append(edges, node.Name+" -> "+dependency.Name)
		}
	}
	slices.Sort(edges)
	return edges
}
