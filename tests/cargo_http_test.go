package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"kf2/src/cargo"
)

// TestCargoIndexPaths проверяет пути Cargo и отсутствие запросов транзитивных пакетов.
func TestCargoIndexPaths(t *testing.T) {
	cases := map[string]string{
		"A": "/index/1/a", "ab": "/index/2/ab", "AbC": "/index/3/a/abc",
		"abcd": "/index/ab/cd/abcd", "foo_bar": "/index/fo/o_/foo_bar",
		"foo-bar": "/index/fo/o-/foo-bar", "serde": "/index/se/rd/serde",
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			body := indexLine(t, map[string]any{"name": name})
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				requests.Add(1)
				if request.URL.Path != path || request.Method != http.MethodGet {
					t.Errorf("неожиданный запрос: %s %s", request.Method, request.URL.Path)
				}
				_, _ = writer.Write([]byte(body))
			}))
			defer server.Close()
			for _, base := range []string{"/index", "/index/"} {
				if _, err := (cargo.Client{}).Fetch(context.Background(), server.URL+base, name); err != nil {
					t.Fatal(err)
				}
			}
			const expectedRequests = 2
			if requests.Load() != expectedRequests {
				t.Fatalf("неожиданное количество запросов: %d", requests.Load())
			}
		})
	}
}

// TestCargoHTTPStatus проверяет сообщения об HTTP-ошибках.
func TestCargoHTTPStatus(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusGone, http.StatusForbidden,
		http.StatusUnavailableForLegalReasons, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(status)
			}))
			defer server.Close()
			_, err := (cargo.Client{}).Fetch(context.Background(), server.URL, "serde")
			if err == nil || !strings.Contains(err.Error(), "HTTP") {
				t.Fatalf("ожидалась ошибка HTTP, получено %v", err)
			}
		})
	}
}

// TestCargoConnectionAndCancellation проверяет сетевой сбой и отменённый контекст.
func TestCargoConnectionAndCancellation(t *testing.T) {
	server := serveIndex(t, "")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := (cargo.Client{}).Fetch(ctx, server.URL, "serde"); !errors.Is(err, context.Canceled) {
		t.Fatalf("ожидалась отмена запроса, получено %v", err)
	}
	server.Close()
	if _, err := (cargo.Client{}).Fetch(context.Background(), server.URL, "serde"); err == nil {
		t.Fatal("ожидалась ошибка подключения")
	}
}

// TestCargoTimeout проверяет ограничение времени HTTP-запроса.
func TestCargoTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()
	const timeout = 50 * time.Millisecond
	client := cargo.Client{HTTP: &http.Client{Timeout: timeout}}
	_, err := client.Fetch(context.Background(), server.URL, "serde")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ожидался тайм-аут, получено %v", err)
	}
}

// TestCargoTruncatedResponse проверяет сбой при чтении неполного тела ответа.
func TestCargoTruncatedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Length", "100")
		_, _ = writer.Write([]byte("{}"))
	}))
	defer server.Close()
	_, err := (cargo.Client{}).Fetch(context.Background(), server.URL, "serde")
	if err == nil || !strings.Contains(err.Error(), "чтение ответа") {
		t.Fatalf("ожидалась ошибка чтения, получено %v", err)
	}
}

// TestCargoIndexLimits проверяет ограничения размера ответа и одной строки.
func TestCargoIndexLimits(t *testing.T) {
	const (
		oversizedLine = 4*1024*1024 + 1
		oversizedBody = 32*1024*1024 + 1
	)
	for _, size := range []int{oversizedLine, oversizedBody} {
		if _, err := fetchIndex(t, strings.Repeat(" ", size)); err == nil {
			t.Fatal("ожидалась ошибка размера индекса")
		}
	}
}

// TestCargoInvalidAddresses проверяет URL и имена до сетевого запроса.
func TestCargoInvalidAddresses(t *testing.T) {
	cases := [][2]string{
		{"https://[", "serde"}, {"ftp://example.org", "serde"}, {"https:///", "serde"},
		{"https://example.org?query=1", "serde"}, {"https://example.org#fragment", "serde"},
		{"https://example.org", ""}, {"https://example.org", "../serde"},
	}
	for _, test := range cases {
		if _, err := (cargo.Client{}).Fetch(context.Background(), test[0], test[1]); err == nil {
			t.Fatalf("ожидалась ошибка для %v", test)
		}
	}
}
