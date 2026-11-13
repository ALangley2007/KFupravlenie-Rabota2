package cargo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	maxIndexLineBytes = 4 * 1024 * 1024
	initialLineBytes  = 64 * 1024
	maxIndexSchema    = 2
	firstIndexLine    = 1
)

type indexEntry struct {
	Name         string       `json:"name"`
	Version      string       `json:"vers"`
	Dependencies []Dependency `json:"deps"`
	Yanked       *bool        `json:"yanked"`
	Schema       uint32       `json:"v"`
}

func parseIndex(data []byte, name string) (Release, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, initialLineBytes), maxIndexLineBytes)
	var best indexEntry
	var bestVersion version
	line := firstIndexLine
	for ; scanner.Scan(); line++ {
		entry, parsed, err := parseEntry(scanner.Bytes(), name)
		if err != nil {
			return Release{}, fmt.Errorf("repository: строка %d индекса: %w", line, err)
		}
		if entry.Yanked == nil || *entry.Yanked {
			continue
		}
		if best.Version == "" || preferVersion(parsed, bestVersion) {
			best, bestVersion = entry, parsed
		}
	}
	if err := scanner.Err(); err != nil {
		return Release{}, fmt.Errorf("repository: чтение индекса: %w", err)
	}
	if best.Version == "" {
		return Release{}, fmt.Errorf("repository: пакет %s: нет доступных версий", name)
	}
	return Release{Name: best.Name, Version: best.Version, Dependencies: best.Dependencies}, nil
}

func parseEntry(data []byte, name string) (indexEntry, version, error) {
	entry, err := decodeEntry(data)
	if err != nil {
		return entry, version{}, fmt.Errorf("некорректный JSON: %w", err)
	}
	if entry.Schema > maxIndexSchema {
		return indexEntry{}, version{}, nil
	}
	if !strings.EqualFold(entry.Name, name) || !crateName.MatchString(entry.Name) {
		return entry, version{}, fmt.Errorf("ожидался пакет %s, получено %q", name, entry.Name)
	}
	parsed, err := parseVersion(entry.Version)
	if err != nil {
		return entry, version{}, err
	}
	if entry.Dependencies == nil || entry.Yanked == nil {
		return entry, version{}, fmt.Errorf("обязательны поля deps (массив) и yanked (boolean)")
	}
	for i := range entry.Dependencies {
		if err := validateDependency(&entry.Dependencies[i]); err != nil {
			return entry, version{}, err
		}
	}
	return entry, parsed, nil
}

func decodeEntry(data []byte) (indexEntry, error) {
	var header struct {
		Schema uint32 `json:"v"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return indexEntry{}, err
	}
	entry := indexEntry{Schema: header.Schema}
	if header.Schema > maxIndexSchema {
		return entry, nil
	}
	err := json.Unmarshal(data, &entry)
	return entry, err
}

func validateDependency(dependency *Dependency) error {
	if !crateName.MatchString(dependency.Name) || !crateName.MatchString(dependency.PackageName()) {
		return fmt.Errorf("некорректное имя зависимости %q", dependency.Name)
	}
	if strings.TrimSpace(dependency.Requirement) == "" {
		return fmt.Errorf("зависимость %s: отсутствует req", dependency.Name)
	}
	if dependency.Kind == "" {
		dependency.Kind = "normal"
	}
	switch dependency.Kind {
	case "normal", "build", "dev":
		return nil
	default:
		return fmt.Errorf("зависимость %s: неизвестный kind %q", dependency.Name, dependency.Kind)
	}
}
