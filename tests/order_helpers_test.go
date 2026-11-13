package tests

import (
	"context"
	"strings"
	"testing"

	"kf2/src/graph"
)

type orderCase struct {
	name   string
	body   string
	root   string
	filter string
	want   string
	cycle  string
}

func orderCases() []orderCase {
	return []orderCase{
		{name: "leaf", body: "A:", want: "A"},
		{name: "chain", body: "A: B\nB: C\nC:", want: "C B A"},
		{name: "branch", body: "A: B C\nB: D\nC: E\nD: F\nE:\nF:", want: "F D B E C A"},
		{name: "diamond", body: "A: B C\nB: D\nC: D\nD:", want: "D B C A"},
		{name: "cross edge", body: "A: B C\nB: C\nC:", want: "C B A"},
		{name: "sibling order", body: "A: C B\nB:\nC:", want: "C B A"},
		{name: "duplicates", body: "A: B B\nB:", want: "B A"},
		{name: "disconnected", body: "A:\nZ: Y\nY:", want: "A"},
		{name: "other root", body: "A: B\nB: C\nC:", root: "B", want: "C B"},
		{name: "filter branch", body: "A: B C\nB: D\nC:\nD:", filter: "B", want: "C A"},
		{name: "filter shared", body: "A: B C\nB: D\nC: D\nD:", filter: "B", want: "D C A"},
		{name: "filter cycle", body: "A: B\nB: C\nC: A", filter: "C", want: "B A"},
		{name: "filter root", body: "A: B\nB:", filter: "A"},
		{name: "filter case", body: "A: B\nB:", filter: "b", want: "B A"},
		{name: "cycle", body: "A: B\nB: C\nC: A", cycle: "A B C A"},
		{name: "nested cycle", body: "A: D B\nB: C\nC: B\nD:", cycle: "B C B"},
		{name: "self", body: "A: A", cycle: "A A"},
	}
}

func (test orderCase) rootName() string {
	if test.root != "" {
		return test.root
	}
	return "A"
}

func orderGraph(t *testing.T, test orderCase) graph.Graph {
	t.Helper()
	dependencies, err := graph.Build(context.Background(), testSource(t, test.body),
		graph.Ref{Name: test.rootName()}, test.filter)
	if err != nil {
		t.Fatal(err)
	}
	return dependencies
}

func orderNames(nodes []graph.Node) string {
	var names []string
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return strings.Join(names, " ")
}
