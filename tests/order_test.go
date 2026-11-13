package tests

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"kf2/src/graph"
)

// TestLoadOrder проверяет топологический порядок, фильтрацию и замкнутые пути циклов.
func TestLoadOrder(t *testing.T) {
	for _, test := range orderCases() {
		t.Run(test.name, func(t *testing.T) {
			dependencies := orderGraph(t, test)
			before := graphText(dependencies)
			order, err := graph.LoadOrder(dependencies)
			if test.cycle != "" {
				assertOrderCycle(t, order, err, test.cycle)
				return
			}
			if err != nil || orderNames(order) != test.want {
				t.Fatalf("порядок %q, ошибка %v; ожидалось %q", orderNames(order), err, test.want)
			}
			if graphText(dependencies) != before {
				t.Fatal("исходный граф изменился")
			}
			assertDependencyOrder(t, order)
		})
	}
}

func assertOrderCycle(t *testing.T, order []graph.Node, err error, want string) {
	t.Helper()
	var cycle *graph.CycleError
	if order != nil || !errors.As(err, &cycle) {
		t.Fatalf("ожидались nil и CycleError, получено %+v, %v", order, err)
	}
	var names []string
	for _, ref := range cycle.Path {
		names = append(names, ref.Name)
	}
	if strings.Join(names, " ") != want || !strings.Contains(err.Error(), strings.Join(names, " -> ")) {
		t.Fatalf("неверный путь цикла: %v", err)
	}
}

func assertDependencyOrder(t *testing.T, order []graph.Node) {
	t.Helper()
	loaded := make(map[graph.Ref]bool)
	for _, node := range order {
		if loaded[node.Ref] {
			t.Fatalf("пакет загружен повторно: %+v", node.Ref)
		}
		for _, dependency := range node.Dependencies {
			if !loaded[dependency] {
				t.Fatalf("%+v выведен до зависимости %+v", node.Ref, dependency)
			}
		}
		loaded[node.Ref] = true
	}
}

// TestLoadOrderRegistries проверяет сохранение версий и различение одинаковых имён разных реестров.
func TestLoadOrderRegistries(t *testing.T) {
	root := graph.Ref{Name: "root", Repository: "https://one.example"}
	first := graph.Ref{Name: "shared", Repository: root.Repository}
	second := graph.Ref{Name: "shared", Repository: "https://two.example"}
	dependencies := graph.Graph{Root: root, Nodes: []graph.Node{
		{Ref: root, Version: "2.0.0", Dependencies: []graph.Ref{first, second}},
		{Ref: first, Version: "1.0.0"}, {Ref: second, Version: "3.0.0"},
	}}
	order, err := graph.LoadOrder(dependencies)
	want := []graph.Node{dependencies.Nodes[1], dependencies.Nodes[2], dependencies.Nodes[0]}
	if err != nil || !reflect.DeepEqual(order, want) {
		t.Fatalf("порядок %+v, ошибка %v; ожидалось %+v", order, err, want)
	}
	assertDependencyOrder(t, order)
}

// TestLoadOrderMissingNode отклоняет неполный граф вместо вывода несуществующей зависимости.
func TestLoadOrderMissingNode(t *testing.T) {
	root, missing := graph.Ref{Name: "A"}, graph.Ref{Name: "B"}
	for _, ref := range []graph.Ref{root, missing} {
		dependencies := graph.Graph{Root: ref, Nodes: []graph.Node{{Ref: root, Dependencies: []graph.Ref{missing}}}}
		order, err := graph.LoadOrder(dependencies)
		if order != nil || err == nil || !strings.Contains(err.Error(), "отсутствует в графе") {
			t.Fatalf("ожидалась ошибка неполного графа, получено %+v, %v", order, err)
		}
	}
}
