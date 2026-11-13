package tests

import (
	"fmt"
	"strings"
	"testing"
)

func assertDOTTopology(t *testing.T, dot, adjacency string) {
	t.Helper()
	if adjacency == "" {
		if strings.Contains(dot, "[label=") || !strings.Contains(dot, "корневой пакет исключён фильтром") {
			t.Fatalf("пустой граф содержит вершины или не объяснён: %s", dot)
		}
		return
	}
	rows := strings.Split(adjacency, "\n")
	ids := make(map[string]string)
	for index, row := range rows {
		name, _, _ := strings.Cut(row, ":")
		ids[name] = fmt.Sprintf("n%d", index)
	}
	var statements []string
	for _, row := range rows {
		name, targets, _ := strings.Cut(row, ":")
		statements = append(statements, fmt.Sprintf("  %s [label=%q];", ids[name], name))
		if targets == "" {
			continue
		}
		for _, target := range strings.Split(targets, ",") {
			statements = append(statements, fmt.Sprintf("  %s -> %s;", ids[name], ids[target]))
		}
	}
	assertDOTStatements(t, dot, statements)
}

func assertDOTStatements(t *testing.T, dot string, statements []string) {
	t.Helper()
	var actual []string
	for line := range strings.SplitSeq(dot, "\n") {
		if strings.Contains(line, "[label=") || strings.Contains(line, " -> ") {
			actual = append(actual, line)
		}
	}
	if len(actual) != len(statements) {
		t.Fatalf("получено %d вершин/рёбер, ожидалось %d: %s", len(actual), len(statements), dot)
	}
	for _, statement := range statements {
		if !strings.Contains(dot, statement+"\n") {
			t.Fatalf("отсутствует %q: %s", statement, dot)
		}
	}
}
