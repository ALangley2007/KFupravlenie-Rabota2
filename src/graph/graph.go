// Package graph строит граф зависимостей рекурсивным обходом в ширину.
package graph

import "context"

// Ref определяет пакет по имени и репозиторию; одинаковые имена разных реестров различаются.
type Ref struct {
	Name       string
	Repository string
}

// Node содержит выбранную версию пакета и ссылки на его прямые зависимости.
type Node struct {
	Ref
	Version      string
	Dependencies []Ref
}

// Source предоставляет прямые зависимости одного пакета.
type Source interface {
	Load(context.Context, Ref) (Node, error)
}

// Graph хранит вершины в порядке BFS; рёбра направлены от пакета к его зависимостям.
type Graph struct {
	Root  Ref
	Nodes []Node
}
