package cargo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	requestTimeout   = 30 * time.Second
	maxResponseBytes = 32 * 1024 * 1024
	oneLetterName    = 1
	twoLetterName    = 2
	threeLetterName  = 3
	firstShardEnd    = 2
	secondShardEnd   = 4
)

var crateName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func indexURL(repository, name string) (string, error) {
	if !crateName.MatchString(name) {
		return "", fmt.Errorf("package: недопустимое имя пакета Cargo %q", name)
	}
	address, err := url.Parse(repository)
	if err != nil {
		return "", fmt.Errorf("repository: некорректный URL: %w", err)
	}
	if (address.Scheme != "http" && address.Scheme != "https") || address.Hostname() == "" {
		return "", fmt.Errorf("repository: требуется HTTP(S) URL корня sparse-индекса Cargo")
	}
	if address.RawQuery != "" || address.Fragment != "" {
		return "", fmt.Errorf("repository: URL sparse-индекса не должен содержать query или fragment")
	}
	return url.JoinPath(repository, indexPath(strings.ToLower(name)))
}

func indexPath(name string) string {
	switch len(name) {
	case oneLetterName:
		return "1/" + name
	case twoLetterName:
		return "2/" + name
	case threeLetterName:
		return "3/" + name[:oneLetterName] + "/" + name
	default:
		return name[:firstShardEnd] + "/" + name[firstShardEnd:secondShardEnd] + "/" + name
	}
}

func (client Client) download(ctx context.Context, address, name string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, fmt.Errorf("repository: HTTP-запрос: %w", err)
	}
	request.Header.Set("Accept", "text/plain")
	request.Header.Set("User-Agent", "kf2/2")
	httpClient := client.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: requestTimeout}
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("repository: загрузка пакета %s: %w", name, err)
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository: пакет %s не найден (HTTP 404)", name)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("repository: пакет %s: HTTP %d", name, response.StatusCode)
	}
	return readResponse(response.Body)
}

func readResponse(reader io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("repository: чтение ответа: %w", err)
	}
	if len(data) > maxResponseBytes {
		return nil, fmt.Errorf("repository: ответ превышает %d байт", maxResponseBytes)
	}
	return data, nil
}
