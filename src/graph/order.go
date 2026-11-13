package graph

import (
	"fmt"
	"slices"
	"strings"
)

type visitState uint8

const (
	unvisited visitState = iota
	active
	complete
	emptyOrder = 0
)

// CycleError содержит замкнутый путь цикла, препятствующего порядку загрузки.
type CycleError struct {
	Path []Ref
}

// Error сообщает, почему нельзя загрузить каждую зависимость раньше её потребителя.
func (err *CycleError) Error() string {
	var names []string
	for _, ref := range err.Path {
		name := ref.Name
		if ref.Repository != "" {
			name += " [" + ref.Repository + "]"
		}
		names = append(names, name)
	}
	return "порядок загрузки невозможен: цикл " + strings.Join(names, " -> ")
}

type loadPlan struct {
	nodes  map[Ref]Node
	states map[Ref]visitState
	path   []Ref
	order  []Node
}

// LoadOrder возвращает достижимые пакеты в порядке «зависимости перед потребителем», включая корень последним.
// Порядок завершения DFS детерминирован порядком рёбер; общие зависимости выводятся один раз.
// При цикле возвращаются nil и CycleError; для пустого графа возвращается пустой порядок.
func LoadOrder(dependencies Graph) ([]Node, error) {
	if len(dependencies.Nodes) == emptyOrder {
		return nil, nil
	}
	plan := loadPlan{nodes: make(map[Ref]Node), states: make(map[Ref]visitState)}
	for _, node := range dependencies.Nodes {
		plan.nodes[node.Ref] = node
	}
	if err := plan.visit(dependencies.Root); err != nil {
		return nil, err
	}
	return plan.order, nil
}

func (plan *loadPlan) visit(ref Ref) error {
	switch plan.states[ref] {
	case complete:
		return nil
	case active:
		path := append([]Ref(nil), plan.path[slices.Index(plan.path, ref):]...)
		return &CycleError{Path: append(path, ref)}
	}
	node, ok := plan.nodes[ref]
	if !ok {
		return fmt.Errorf("порядок загрузки: пакет %s [%s] отсутствует в графе", ref.Name, ref.Repository)
	}
	plan.states[ref] = active
	plan.path = append(plan.path, ref)
	for _, dependency := range node.Dependencies {
		if err := plan.visit(dependency); err != nil {
			return err
		}
	}
	plan.path = plan.path[:len(plan.path)-1]
	plan.states[ref] = complete
	plan.order = append(plan.order, node)
	return nil
}
