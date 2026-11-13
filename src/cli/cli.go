// Package cli реализует интерфейс командной строки анализатора зависимостей Cargo.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"kf2/src/cargo"
	"kf2/src/config"
)

const (
	exitSuccess = 0
	exitFailure = 1
	noArguments = 0
)

// Run получает прямые зависимости пакета из XML-конфигурации и возвращает код завершения.
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
	release, err := fetchRelease(cfg)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка:", err)
		return exitFailure
	}
	err = writeRelease(stdout, release)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка вывода:", err)
		return exitFailure
	}
	return exitSuccess
}

func fetchRelease(cfg config.Config) (cargo.Release, error) {
	if cfg.TestMode {
		return cargo.Release{}, fmt.Errorf("test_mode: чтение тестового графа предусмотрено этапом 3")
	}
	return (cargo.Client{}).Fetch(context.Background(), cfg.Repository, cfg.Package)
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
