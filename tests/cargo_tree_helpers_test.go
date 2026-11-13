package tests

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"kf2/src/graph"
)

const (
	cargoTreeTimeout = 3 * time.Minute
	rootDepth        = 0
)

var cargoTreeLine = regexp.MustCompile(`^(\d+)([A-Za-z0-9_-]+) v([^ ]+)`)

func cargoTreeManifest(t *testing.T, root graph.Node) string {
	t.Helper()
	directory := t.TempDir()
	manifest := "[package]\nname = \"kf2-visual-comparison\"\nversion = \"0.1.0\"\nedition = \"2021\"\n"
	manifest += fmt.Sprintf("\n[dependencies]\n%s = \"=%s\"\n", root.Name, root.Version)
	writeFixtureFile(t, filepath.Join(directory, "lib.rs"), "")
	manifest += "\n[lib]\npath = \"lib.rs\"\n"
	path := filepath.Join(directory, "Cargo.toml")
	writeFixtureFile(t, path, manifest)
	return path
}

func runCargoTree(t *testing.T, root graph.Node, manifest, cache, target, prefix string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cargoTreeTimeout)
	defer cancel()
	args := []string{"tree", "--manifest-path", manifest, "-p", root.Name + "@" + root.Version,
		"--prefix", prefix, "--no-dedupe", "--edges", "normal,build,dev", "--color", "never"}
	if target != "" {
		args = append(args, "--target", target)
	}
	command := exec.CommandContext(ctx, "cargo", args...)
	command.Env = append(os.Environ(), "CARGO_HOME="+cache, "CARGO_TARGET_DIR="+t.TempDir())
	var errors strings.Builder
	command.Stderr = &errors
	output, err := command.Output()
	if err != nil {
		t.Fatalf("cargo tree: %v\n%s", err, errors.String())
	}
	t.Logf("cargo %s\n%s", strings.Join(args, " "), output)
	return string(output)
}

func parseCargoTree(t *testing.T, output string) []string {
	t.Helper()
	edges := []string{}
	var path []string
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		fields := cargoTreeFields(t, line)
		depth, err := strconv.Atoi(fields[1])
		if err != nil || depth > len(path) {
			t.Fatalf("неверная глубина дерева: %q: %v", line, err)
		}
		path = path[:depth]
		if depth != rootDepth {
			edges = append(edges, path[depth-1]+" -> "+fields[2])
		}
		path = append(path, fields[2])
	}
	slices.Sort(edges)
	return slices.Compact(edges)
}

func cargoTreePackages(t *testing.T, output string) map[string]string {
	t.Helper()
	names := make(map[string]string)
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		fields := cargoTreeFields(t, line)
		names[fields[2]] = fields[3]
	}
	return names
}

func cargoTreeFields(t *testing.T, line string) []string {
	t.Helper()
	fields := cargoTreeLine.FindStringSubmatch(line)
	if fields == nil {
		t.Fatalf("неизвестная строка cargo tree: %q", line)
	}
	return fields
}
