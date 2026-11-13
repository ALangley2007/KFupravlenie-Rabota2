package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	minPort = 1
	maxPort = 65535
)

var (
	packageName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	testName    = regexp.MustCompile(`^[A-Z]$`)
)

func (cfg Config) validate(baseDir string) error {
	if !packageName.MatchString(cfg.Package) {
		return fmt.Errorf("package: нужно непустое имя из A–Z, a–z, 0–9, '-' и '_'")
	}
	if cfg.TestMode && !testName.MatchString(cfg.Package) {
		return fmt.Errorf("package: в тестовом режиме нужна одна буква A–Z")
	}
	if strings.ContainsFunc(cfg.Filter, unicode.IsControl) {
		return fmt.Errorf("filter: управляющие символы недопустимы")
	}
	if cfg.Repository == "" {
		return fmt.Errorf("repository: значение не должно быть пустым")
	}
	if strings.ContainsFunc(cfg.Repository, unicode.IsControl) {
		return fmt.Errorf("repository: управляющие символы недопустимы")
	}
	if cfg.TestMode {
		return validateFile(cfg.Repository, baseDir)
	}
	return validateURL(cfg.Repository)
}

func validateURL(value string) error {
	address, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("repository: некорректный URL: %w", err)
	}
	if address.Scheme != "http" && address.Scheme != "https" {
		return fmt.Errorf("repository: требуется абсолютный URL с протоколом http или https")
	}
	if address.Hostname() == "" || strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("repository: нужен узел URL; пробелы недопустимы")
	}
	if strings.HasSuffix(address.Host, ":") {
		return fmt.Errorf("repository: порт URL не должен быть пустым")
	}
	return validatePort(address.Port())
}

func validatePort(value string) error {
	if value == "" {
		return nil
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < minPort || port > maxPort {
		return fmt.Errorf("repository: порт URL должен быть в диапазоне %d–%d", minPort, maxPort)
	}
	return nil
}

func validateFile(value, baseDir string) error {
	if strings.Contains(value, "://") {
		return fmt.Errorf("repository: в тестовом режиме нужен локальный файл")
	}
	path := value
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("repository: файл тестового репозитория: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("repository: путь должен указывать на обычный файл")
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("repository: файл недоступен для чтения: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("repository: закрытие файла: %w", err)
	}
	return nil
}
