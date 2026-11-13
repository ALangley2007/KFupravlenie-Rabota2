// Package cli реализует интерфейс командной строки прототипа.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"kf2/src/config"
)

const (
	exitSuccess = 0
	exitFailure = 1
	noArguments = 0
)

// Run выводит параметры из XML и возвращает код завершения.
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
	_, err = fmt.Fprintf(stdout, "package=%s\nrepository=%s\ntest_mode=%t\nfilter=%s\n",
		cfg.Package, cfg.Repository, cfg.TestMode, cfg.Filter)
	if err != nil {
		fmt.Fprintln(stderr, "Ошибка вывода:", err)
		return exitFailure
	}
	return exitSuccess
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
