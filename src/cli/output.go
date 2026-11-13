package cli

import (
	"fmt"
	"io"
	"strings"

	"kf2/src/graph"
)

const noNodes = 0

func writeGraph(writer io.Writer, dependencies graph.Graph) error {
	var output strings.Builder
	fmt.Fprintf(&output, "Пакет: %s\nГраф зависимостей (BFS):\n", label(dependencies.Root))
	if len(dependencies.Nodes) == noNodes {
		output.WriteString("(пуст: корневой пакет исключён фильтром)\n")
	}
	for _, node := range dependencies.Nodes {
		name := label(node.Ref)
		if node.Version != "" {
			name += "@" + node.Version
		}
		var targets []string
		for _, dependency := range node.Dependencies {
			targets = append(targets, label(dependency))
		}
		if len(targets) == noNodes {
			targets = append(targets, "(нет)")
		}
		fmt.Fprintf(&output, "%s -> %s\n", name, strings.Join(targets, ", "))
	}
	_, err := io.WriteString(writer, output.String())
	return err
}

func label(ref graph.Ref) string {
	if ref.Repository != "" {
		return fmt.Sprintf("%s [%s]", ref.Name, ref.Repository)
	}
	return ref.Name
}
