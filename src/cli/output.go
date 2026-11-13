package cli

import (
	"fmt"
	"io"
	"strings"

	"kf2/src/cargo"
)

const noDependencies = 0

func writeRelease(writer io.Writer, release cargo.Release) error {
	var output strings.Builder
	fmt.Fprintf(&output, "Пакет: %s@%s\nПрямые зависимости:\n", release.Name, release.Version)
	if len(release.Dependencies) == noDependencies {
		output.WriteString("(нет)\n")
	}
	for _, dependency := range release.Dependencies {
		fmt.Fprintf(&output, "- %s %s [kind=%s optional=%t",
			dependency.PackageName(), dependency.Requirement, dependency.Kind, dependency.Optional)
		if dependency.Package != "" {
			fmt.Fprintf(&output, " alias=%s", dependency.Name)
		}
		if dependency.Target != "" {
			fmt.Fprintf(&output, " target=%q", dependency.Target)
		}
		if dependency.Registry != "" {
			fmt.Fprintf(&output, " registry=%q", dependency.Registry)
		}
		output.WriteString("]\n")
	}
	_, err := io.WriteString(writer, output.String())
	return err
}
