package graph

import (
	"fmt"
	"strings"
)

const noDOTNodes = 0

var dotEscapes = strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", "\\n", "\r", "\\r")

// DOT возвращает описание построенного графа на языке Graphviz без обращения к внешним программам.
// Идентификаторы назначаются в порядке BFS; подписи сохраняют версии и различают реестры.
// Все вершины объявляются явно, включая листья; циклы и самозависимости сохраняются.
func DOT(dependencies Graph) string {
	var output strings.Builder
	output.WriteString("digraph dependencies {\n  rankdir=LR;\n  node [shape=box];\n")
	if len(dependencies.Nodes) == noDOTNodes {
		output.WriteString("  label=\"Пустой граф: корневой пакет исключён фильтром\";\n")
	}
	identifiers := make(map[Ref]string)
	for index, node := range dependencies.Nodes {
		id := fmt.Sprintf("n%d", index)
		identifiers[node.Ref] = id
		fmt.Fprintf(&output, "  %s [label=\"%s\"];\n", id, dotEscapes.Replace(dotLabel(node)))
	}
	for _, node := range dependencies.Nodes {
		for _, dependency := range node.Dependencies {
			fmt.Fprintf(&output, "  %s -> %s;\n", identifiers[node.Ref], identifiers[dependency])
		}
	}
	output.WriteString("}\n")
	return output.String()
}

func dotLabel(node Node) string {
	label := node.Name
	if node.Version != "" {
		label += "@" + node.Version
	}
	if node.Repository != "" {
		label += "\n" + node.Repository
	}
	return label
}
