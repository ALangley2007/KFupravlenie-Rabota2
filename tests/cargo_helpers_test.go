package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kf2/src/cargo"
)

func indexLine(t *testing.T, fields map[string]any) string {
	t.Helper()
	entry := map[string]any{
		"name": "serde", "vers": "1.0.0", "deps": []cargo.Dependency{}, "yanked": false,
	}
	for key, value := range fields {
		entry[key] = value
	}
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	return string(data) + "\n"
}

func serveIndex(t *testing.T, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/se/rd/serde" {
			http.NotFound(writer, request)
			return
		}
		_, _ = writer.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func fetchIndex(t *testing.T, body string) (cargo.Release, error) {
	t.Helper()
	server := serveIndex(t, body)
	return (cargo.Client{}).Fetch(context.Background(), server.URL, "serde")
}

func remoteConfig(t *testing.T, body string) string {
	t.Helper()
	server := serveIndex(t, body)
	content := strings.ReplaceAll(validXML, "https://index.crates.io", server.URL)
	return writeConfig(t, t.TempDir(), content)
}
