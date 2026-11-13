package tests

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"kf2/src/graph"
)

const cargoComparisonTimeout = time.Minute

// TestCargoComparison сравнивает порядок с настоящим Cargo; включается командой ./run.sh compare.
func TestCargoComparison(t *testing.T) {
	if os.Getenv("KF2_CARGO_COMPARE") != "true" {
		t.Skip("для сравнения с установленными Cargo и rustc выполните ./run.sh compare")
	}
	version, err := exec.Command("cargo", "--version").CombinedOutput()
	if err != nil {
		t.Fatalf("Cargo недоступен: %v: %s", err, version)
	}
	t.Log(strings.TrimSpace(string(version)))
	for _, test := range orderCases() {
		if test.want == "" && test.cycle == "" {
			continue
		}
		t.Run(test.name, func(t *testing.T) { compareCargoOrder(t, test) })
	}
}

func compareCargoOrder(t *testing.T, test orderCase) {
	t.Helper()
	dependencies := orderGraph(t, test)
	programOrder, orderError := graph.LoadOrder(dependencies)
	if test.cycle != "" {
		assertOrderCycle(t, programOrder, orderError, test.cycle)
	} else if orderError != nil || orderNames(programOrder) != test.want {
		t.Fatalf("kf2: порядок %q, ошибка %v; ожидалось %q", orderNames(programOrder), orderError, test.want)
	}
	stdout, stderr, err := buildCargoFixture(t, dependencies)
	if test.cycle != "" {
		if err == nil || !strings.Contains(stderr, "cyclic package dependency") {
			t.Fatalf("Cargo не обнаружил цикл: %v\n%s", err, stderr)
		}
		t.Logf("Оба инструмента отклоняют цикл %s. Cargo:\n%s", test.cycle, stderr)
		return
	}
	if err != nil {
		t.Fatalf("cargo build: %v\n%s\n%s", err, stdout, stderr)
	}
	order := nodesInCargoOrder(t, dependencies, cargoArtifacts(t, stdout))
	assertDependencyOrder(t, order)
	t.Logf("kf2: %s; Cargo (-j 1): %s", orderNames(programOrder), orderNames(order))
	if orderNames(order) != orderNames(programOrder) {
		t.Log("Различие допустимо: независимые пакеты допускают несколько топологических порядков.")
	}
}

func buildCargoFixture(t *testing.T, dependencies graph.Graph) ([]byte, string, error) {
	t.Helper()
	manifest := cargoFixture(t, dependencies)
	ctx, cancel := context.WithTimeout(context.Background(), cargoComparisonTimeout)
	defer cancel()
	command := exec.CommandContext(ctx, "cargo", "build", "--offline", "-j", "1",
		"--message-format=json", "--color", "never", "--manifest-path", manifest)
	command.Env = append(os.Environ(), "CARGO_HOME="+t.TempDir(), "CARGO_TARGET_DIR="+t.TempDir())
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatal(ctx.Err())
	}
	return stdout.Bytes(), stderr.String(), err
}
