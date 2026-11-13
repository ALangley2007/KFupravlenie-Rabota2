package repository

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"kf2/src/graph"
)

const (
	firstLine   = 1
	noTestNodes = 0
)

var testName = regexp.MustCompile(`^[A-Z]$`)

// Test содержит проверенное описание тестового репозитория с именами A–Z.
type Test struct {
	nodes map[string]graph.Node
	order []string
}

// LoadTest читает строки вида «A: B C»; пустые строки пропускаются, ссылки требуют объявления.
func LoadTest(path string) (*Test, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("repository: открытие тестового графа: %w", err)
	}
	defer file.Close()
	source := &Test{nodes: make(map[string]graph.Node)}
	scanner := bufio.NewScanner(file)
	for line := firstLine; scanner.Scan(); line++ {
		if err := source.parseLine(strings.TrimSpace(scanner.Text())); err != nil {
			return nil, fmt.Errorf("repository: строка %d: %w", line, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("repository: чтение тестового графа: %w", err)
	}
	if err := source.validate(); err != nil {
		return nil, err
	}
	return source, nil
}

// Load возвращает зависимости тестового пакета или ошибку отсутствующего имени.
func (source *Test) Load(ctx context.Context, ref graph.Ref) (graph.Node, error) {
	if err := ctx.Err(); err != nil {
		return graph.Node{}, err
	}
	node, ok := source.nodes[ref.Name]
	if !ok {
		return graph.Node{}, fmt.Errorf("repository: пакет %s не найден в тестовом графе", ref.Name)
	}
	return node, nil
}

func (source *Test) parseLine(line string) error {
	if line == "" {
		return nil
	}
	name, dependencies, ok := strings.Cut(line, ":")
	name = strings.TrimSpace(name)
	if !ok || !testName.MatchString(name) {
		return fmt.Errorf("ожидается строка A: B C с именами из одной буквы A–Z")
	}
	if _, exists := source.nodes[name]; exists {
		return fmt.Errorf("повторное объявление пакета %s", name)
	}
	node := graph.Node{Ref: graph.Ref{Name: name}}
	for _, dependency := range strings.Fields(dependencies) {
		if !testName.MatchString(dependency) {
			return fmt.Errorf("недопустимое имя зависимости %q: нужна одна буква A–Z", dependency)
		}
		node.Dependencies = append(node.Dependencies, graph.Ref{Name: dependency})
	}
	source.nodes[name] = node
	source.order = append(source.order, name)
	return nil
}

func (source *Test) validate() error {
	if len(source.nodes) == noTestNodes {
		return fmt.Errorf("repository: тестовый граф пуст")
	}
	for _, name := range source.order {
		for _, dependency := range source.nodes[name].Dependencies {
			if _, ok := source.nodes[dependency.Name]; !ok {
				return fmt.Errorf("repository: пакет %s: зависимость %s не объявлена", name, dependency.Name)
			}
		}
	}
	return nil
}
