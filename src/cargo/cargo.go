// Package cargo читает прямые зависимости из HTTP sparse-индекса Cargo без менеджера пакетов.
package cargo

import (
	"context"
	"net/http"
)

// Dependency описывает объявленную прямую зависимость, включая псевдоним и ограничения платформы.
type Dependency struct {
	Name        string `json:"name"`
	Package     string `json:"package"`
	Requirement string `json:"req"`
	Kind        string `json:"kind"`
	Optional    bool   `json:"optional"`
	Target      string `json:"target"`
	Registry    string `json:"registry"`
}

// PackageName возвращает исходное имя пакета с учётом переименования зависимости.
func (dependency Dependency) PackageName() string {
	if dependency.Package != "" {
		return dependency.Package
	}
	return dependency.Name
}

// Release содержит выбранную версию пакета и все объявленные прямые зависимости.
type Release struct {
	Name         string
	Version      string
	Dependencies []Dependency
}

// Client получает данные по HTTP; нулевое значение использует тайм-аут 30 секунд.
type Client struct {
	// HTTP позволяет задать транспорт и тайм-аут; nil использует стандартный клиент.
	HTTP *http.Client
}

// Fetch выбирает наибольшую стабильную SemVer-версию, не помеченную yanked.
// Если стабильных версий нет, выбирается наибольшая предварительная версия.
func (client Client) Fetch(ctx context.Context, repository, name string) (Release, error) {
	address, err := indexURL(repository, name)
	if err != nil {
		return Release{}, err
	}
	data, err := client.download(ctx, address, name)
	if err != nil {
		return Release{}, err
	}
	return parseIndex(data, name)
}
