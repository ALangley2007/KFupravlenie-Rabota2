package cli

import (
	"fmt"
	"io"

	"kf2/src/graph"
)

const noNodes = 0

func writeGraph(writer io.Writer, dependencies graph.Graph) error {
	_, err := io.WriteString(writer, graph.DOT(dependencies))
	return err
}

func label(ref graph.Ref) string {
	if ref.Repository != "" {
		return fmt.Sprintf("%s [%s]", ref.Name, ref.Repository)
	}
	return ref.Name
}
