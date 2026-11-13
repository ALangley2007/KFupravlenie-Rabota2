package tests

import (
	"strings"
	"testing"

	"kf2/src/graph"
)

// TestDOTTopology проверяет листья, общие зависимости, петли, фильтрацию и порядок BFS в DOT.
func TestDOTTopology(t *testing.T) {
	for _, test := range graphCases() {
		t.Run(test.name, func(t *testing.T) {
			dependencies := orderGraph(t, orderCase{body: test.body, filter: test.filter})
			got := graph.DOT(dependencies)
			if !strings.HasPrefix(got, "digraph dependencies {\n") || !strings.HasSuffix(got, "}\n") {
				t.Fatalf("некорректная оболочка DOT: %s", got)
			}
			assertDOTTopology(t, got, test.want)
			if graph.DOT(dependencies) != got {
				t.Fatal("порядок вывода нестабилен")
			}
		})
	}
}

// TestDOTEscaping проверяет кавычки, обратные слеши, переносы и разные реестры одинаковых имён.
func TestDOTEscaping(t *testing.T) {
	first := graph.Ref{Name: "same", Repository: "https://one.test/\";\\N"}
	second := graph.Ref{Name: "same", Repository: "https://two.test/<index>&"}
	dependencies := graph.Graph{Root: first, Nodes: []graph.Node{
		{Ref: first, Version: "1.2.3", Dependencies: []graph.Ref{second}},
		{Ref: second, Version: "2.0.0"},
	}}
	want := "digraph dependencies {\n  rankdir=LR;\n  node [shape=box];\n" +
		`  n0 [label="same@1.2.3\nhttps://one.test/\";\\N"];` + "\n" +
		`  n1 [label="same@2.0.0\nhttps://two.test/<index>&"];` + "\n  n0 -> n1;\n}\n"
	if got := graph.DOT(dependencies); got != want {
		t.Fatalf("DOT=%q; ожидалось %q", got, want)
	}
}
