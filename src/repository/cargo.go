// Package repository предоставляет источники графа: индекс Cargo и локальный тестовый файл.
package repository

import (
	"context"
	"strings"

	"kf2/src/cargo"
	"kf2/src/graph"
)

// Cargo получает прямые зависимости из HTTP sparse-индекса стандартным клиентом Go.
type Cargo struct {
	Client cargo.Client
}

// CargoRef приводит имя и URL к единому виду для обнаружения циклов и общих зависимостей.
func CargoRef(repository, name string) graph.Ref {
	repository = strings.TrimPrefix(repository, "sparse+")
	repository = strings.TrimRight(repository, "/")
	return graph.Ref{Name: strings.ToLower(name), Repository: repository}
}

// Load выбирает версию по правилам этапа 2; ограничения req не разрешаются как Cargo.lock.
func (source Cargo) Load(ctx context.Context, ref graph.Ref) (graph.Node, error) {
	release, err := source.Client.Fetch(ctx, ref.Repository, ref.Name)
	if err != nil {
		return graph.Node{}, err
	}
	node := graph.Node{Ref: ref, Version: release.Version}
	for _, dependency := range release.Dependencies {
		repository := ref.Repository
		if dependency.Registry != "" {
			repository = dependency.Registry
		}
		node.Dependencies = append(node.Dependencies, CargoRef(repository, dependency.PackageName()))
	}
	return node, nil
}
