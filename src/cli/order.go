package cli

import (
	"fmt"
	"io"
	"strings"

	"kf2/src/graph"
)

func writeLoadOrder(writer io.Writer, dependencies graph.Graph) error {
	order, err := graph.LoadOrder(dependencies)
	if err != nil {
		return err
	}
	var output strings.Builder
	fmt.Fprintf(&output, "Пакет: %s\nПорядок загрузки (зависимости перед пакетом):\n", label(dependencies.Root))
	if len(order) == noNodes {
		output.WriteString("(пуст: корневой пакет исключён фильтром)\n")
	}
	for _, node := range order {
		name := label(node.Ref)
		if node.Version != "" {
			name += "@" + node.Version
		}
		fmt.Fprintln(&output, name)
	}
	_, err = io.WriteString(writer, output.String())
	return err
}
