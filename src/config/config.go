// Package config загружает и проверяет пользовательские параметры из XML.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config содержит параметры анализа, заданные в XML-файле.
type Config struct {
	// Package задаёт имя анализируемого пакета.
	Package string
	// Repository содержит URL или путь в исходном виде из конфигурации.
	Repository string
	// TestMode включает использование локального файла репозитория.
	TestMode bool
	// Filter задаёт подстроку исключения пакетов.
	Filter string
	// LoadOrder включает вывод зависимостей перед зависящими от них пакетами.
	LoadOrder bool
}

// Load читает конфигурацию; пути отсчитываются от каталога XML.
func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("файл конфигурации: %w", err)
	}
	defer file.Close()
	values, err := decode(file)
	if err != nil {
		return Config{}, fmt.Errorf("конфигурация XML: %w", err)
	}
	cfg, err := fromValues(values)
	if err != nil {
		return Config{}, err
	}
	if err := cfg.validate(filepath.Dir(path)); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func fromValues(values map[string]string) (Config, error) {
	mode := strings.TrimSpace(values["test_mode"])
	if mode != "true" && mode != "false" {
		return Config{}, fmt.Errorf("test_mode: ожидается true или false")
	}
	order, present := values["load_order"]
	order = strings.TrimSpace(order)
	if present && order != "true" && order != "false" {
		return Config{}, fmt.Errorf("load_order: ожидается true или false")
	}
	return Config{
		Package:    strings.TrimSpace(values["package"]),
		Repository: strings.TrimSpace(values["repository"]),
		TestMode:   mode == "true",
		Filter:     values["filter"],
		LoadOrder:  order == "true",
	}, nil
}
