package graph

import (
	"context"
	"fmt"
	"strings"
)

const emptyFrontier = 0

type traversal struct {
	source Source
	filter string
	seen   map[Ref]bool
	graph  Graph
}

// Build обходит граф по уровням BFS, рекурсивно переходя к следующему уровню.
// Фильтр регистрозависим; пустая подстрока ничего не исключает, непустая исключает и корень.
// Вершины посещаются один раз, а рёбра циклов и общих зависимостей сохраняются.
func Build(ctx context.Context, source Source, root Ref, filter string) (Graph, error) {
	walk := traversal{source: source, filter: filter, seen: make(map[Ref]bool), graph: Graph{Root: root}}
	if err := ctx.Err(); err != nil {
		return Graph{}, err
	}
	if walk.excluded(root) {
		return walk.graph, nil
	}
	walk.seen[root] = true
	if err := walk.level(ctx, []Ref{root}); err != nil {
		return Graph{}, err
	}
	return walk.graph, nil
}

func (walk *traversal) level(ctx context.Context, frontier []Ref) error {
	if len(frontier) == emptyFrontier {
		return nil
	}
	var next []Ref
	for _, ref := range frontier {
		if err := ctx.Err(); err != nil {
			return err
		}
		node, err := walk.source.Load(ctx, ref)
		if err != nil {
			return fmt.Errorf("пакет %s: %w", ref.Name, err)
		}
		node.Ref = ref
		node.Dependencies = walk.dependencies(node.Dependencies, &next)
		walk.graph.Nodes = append(walk.graph.Nodes, node)
	}
	return walk.level(ctx, next)
}

func (walk *traversal) dependencies(dependencies []Ref, next *[]Ref) []Ref {
	var included []Ref
	edges := make(map[Ref]bool)
	for _, ref := range dependencies {
		if walk.excluded(ref) || edges[ref] {
			continue
		}
		edges[ref] = true
		included = append(included, ref)
		if !walk.seen[ref] {
			walk.seen[ref] = true
			*next = append(*next, ref)
		}
	}
	return included
}

func (walk *traversal) excluded(ref Ref) bool {
	return walk.filter != "" && strings.Contains(ref.Name, walk.filter)
}
