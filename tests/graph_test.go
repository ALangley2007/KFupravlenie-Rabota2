package tests

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"kf2/src/graph"
)

type graphCase struct {
	name   string
	body   string
	filter string
	want   string
}

func graphCases() []graphCase {
	return []graphCase{
		{"chain", "A: B\nB: C\nC:", "", "A:B\nB:C\nC:"},
		{"breadth first", "A: B C\nB: D\nC: E\nD: F\nE:\nF:", "", "A:B,C\nB:D\nC:E\nD:F\nE:\nF:"},
		{"diamond", "A: B C\nB: D\nC: D\nD:", "", "A:B,C\nB:D\nC:D\nD:"},
		{"cycle", "A: B\nB: C\nC: A", "", "A:B\nB:C\nC:A"},
		{"self and duplicate", "A: A B B\nB: B", "", "A:A,B\nB:B"},
		{"disconnected", "A:\nZ: Y\nY:", "", "A:"},
		{"filter branch", "A: B C\nB: D\nC:\nD:", "B", "A:C\nC:"},
		{"filter shared", "A: B C\nB: D\nC: D\nD:", "B", "A:C\nC:D\nD:"},
		{"filter transitive", "A: B\nB: C\nC: A", "C", "A:B\nB:"},
		{"filter root", "A: B\nB:", "A", ""},
		{"filter case", "A: B\nB:", "b", "A:B\nB:"},
	}
}

// TestGraphTraversal проверяет рекурсивный BFS, фильтр, циклы и однократную загрузку вершин.
func TestGraphTraversal(t *testing.T) {
	for _, test := range graphCases() {
		t.Run(test.name, func(t *testing.T) {
			source := testSource(t, test.body)
			got, err := graph.Build(context.Background(), source, graph.Ref{Name: "A"}, test.filter)
			if err != nil || graphText(got) != test.want {
				t.Fatalf("граф %q, ошибка %v; ожидалось %q", graphText(got), err, test.want)
			}
			var wantOrder []graph.Ref
			for _, node := range got.Nodes {
				wantOrder = append(wantOrder, node.Ref)
			}
			if !reflect.DeepEqual(source.loaded, wantOrder) {
				t.Fatalf("порядок запросов %+v; ожидалось %+v", source.loaded, wantOrder)
			}
		})
	}
}

// TestGraphCancellation проверяет отмену до обращения к источнику, в том числе с фильтром корня.
func TestGraphCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, filter := range []string{"", "A"} {
		source := testSource(t, "A:")
		_, err := graph.Build(ctx, source, graph.Ref{Name: "A"}, filter)
		if !errors.Is(err, context.Canceled) || len(source.loaded) != noOutput {
			t.Fatalf("ошибка %v, запросы %+v", err, source.loaded)
		}
	}
}

// TestGraphSourceFailure проверяет отсутствие частичного результата при ошибке источника.
func TestGraphSourceFailure(t *testing.T) {
	source := testSource(t, "A:")
	got, err := graph.Build(context.Background(), source, graph.Ref{Name: "Z"}, "")
	if err == nil || len(got.Nodes) != noOutput {
		t.Fatalf("ожидалась ошибка отсутствующего корня, получено %+v, %v", got, err)
	}
}
