// Package cli реализует интерфейс командной строки анализатора зависимостей Cargo.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"kf2/src/config"
	"kf2/src/graph"
	"kf2/src/repository"
)

const (
	exitSuccess = 0
	exitFailure = 1
	noArguments = 0
)

// Run строит граф зависимостей по XML-конфигурации и возвращает код завершения.
func Run(args []string, stdout, stderr io.Writer) int {
	path, err := configPath(args, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitFailure
	}
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitFailure
	}
	dependencies, err := buildGraph(cfg, path)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitFailure
	}
	err = writeGraph(stdout, dependencies)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка вывода:", err)
		return exitFailure
	}
	return exitSuccess
}

func buildGraph(cfg config.Config, configPath string) (graph.Graph, error) {
	var source graph.Source = repository.Cargo{}
	root := repository.CargoRef(cfg.Repository, cfg.Package)
	if cfg.TestMode {
		path := cfg.Repository
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(configPath), path)
		}
		test, err := repository.LoadTest(path)
		if err != nil {
			return graph.Graph{}, err
		}
		source = test
		root = graph.Ref{Name: cfg.Package}
	}
	return graph.Build(context.Background(), source, root, cfg.Filter)
}

func configPath(args []string, stderr io.Writer) (string, error) {
	flags := flag.NewFlagSet("kf2", flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("config", "config.xml", "путь к XML-конфигурации")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	if flags.NArg() != noArguments {
		return "", fmt.Errorf("позиционные аргументы не поддерживаются: %v", flags.Args())
	}
	return *path, nil
}
