package tests

import (
	"reflect"
	"strings"
	"testing"

	"kf2/src/cargo"
)

// TestCargoDependencies проверяет все виды зависимостей, переименования и платформы.
func TestCargoDependencies(t *testing.T) {
	want := []cargo.Dependency{
		{Name: "serde_core", Requirement: "^1.0", Kind: "normal"},
		{Name: "derive", Package: "serde_derive", Requirement: "=1.0.0", Kind: "normal", Optional: true},
		{Name: "cc", Requirement: "^1", Kind: "build"},
		{Name: "test_crate", Requirement: "*", Kind: "dev"},
		{Name: "winapi", Requirement: "^0.3", Kind: "normal", Target: `cfg(windows)`},
		{Name: "winapi", Requirement: "^0.2", Kind: "build", Target: `cfg(unix)`},
		{Name: "custom", Requirement: "^2", Kind: "normal", Registry: "https://registry.example/index"},
	}
	got, err := fetchIndex(t, indexLine(t, map[string]any{"deps": want, "v": 2}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "serde" || got.Version != "1.0.0" || !reflect.DeepEqual(got.Dependencies, want) {
		t.Fatalf("получено %+v; зависимости %+v", got, got.Dependencies)
	}
}

// TestCargoEmptyAndDefaultKind проверяет пустой список и kind по умолчанию.
func TestCargoEmptyAndDefaultKind(t *testing.T) {
	for _, deps := range []any{[]cargo.Dependency{}, []map[string]any{
		{"name": "foo", "req": "^1"}, {"name": "bar", "req": "^2", "kind": nil},
	}} {
		got, err := fetchIndex(t, indexLine(t, map[string]any{"deps": deps}))
		if err != nil {
			t.Fatal(err)
		}
		for _, dependency := range got.Dependencies {
			if dependency.Kind != "normal" {
				t.Fatalf("kind по умолчанию: %+v", dependency)
			}
		}
	}
}

// TestCargoInvalidIndex проверяет ошибки формата индекса и обязательных полей.
func TestCargoInvalidIndex(t *testing.T) {
	cases := []string{
		"", "\n", "{", "null", "[]", "<html>error</html>",
		indexLine(t, map[string]any{"name": "another"}),
		indexLine(t, map[string]any{"deps": nil}),
		indexLine(t, map[string]any{"yanked": nil}),
		indexLine(t, map[string]any{"deps": "invalid"}),
		indexLine(t, map[string]any{"v": -1}),
		indexLine(t, map[string]any{"deps": []map[string]any{{"name": "bad/name", "req": "^1"}}}),
		indexLine(t, map[string]any{"deps": []map[string]any{{"name": "foo"}}}),
		indexLine(t, map[string]any{"deps": []map[string]any{{"name": "foo", "req": " "}}}),
		indexLine(t, map[string]any{"deps": []map[string]any{{"name": "foo", "req": map[string]string{}}}}),
		indexLine(t, map[string]any{"deps": []map[string]any{{"name": "foo", "req": "*", "kind": "x"}}}),
		indexLine(t, map[string]any{"deps": []map[string]any{{"name": "foo", "req": "*", "package": "/"}}}),
		indexLine(t, map[string]any{"vers": "9.0.0", "yanked": true}),
		indexLine(t, map[string]any{"v": 3}),
		`{"name":"serde","vers":"1.0.0","yanked":false}`,
	}
	for _, body := range cases {
		t.Run(body, func(t *testing.T) {
			_, err := fetchIndex(t, body)
			if err == nil || !strings.Contains(err.Error(), "repository:") {
				t.Fatalf("ожидалась ошибка индекса, получено %v", err)
			}
		})
	}
}

// TestCargoSkipsUnavailable проверяет исключение yanked и неподдерживаемых схем.
func TestCargoSkipsUnavailable(t *testing.T) {
	body := indexLine(t, map[string]any{"vers": "9.0.0", "yanked": true})
	body += indexLine(t, map[string]any{"vers": "8.0.0", "v": 3})
	body += indexLine(t, map[string]any{"vers": "1.0.0", "v": 1})
	got, err := fetchIndex(t, body)
	if err != nil || got.Version != "1.0.0" {
		t.Fatalf("получено %+v, ошибка %v", got, err)
	}
}

// TestCargoSkipsUnknownSchemaFields проверяет пропуск новых схем с изменёнными типами полей.
func TestCargoSkipsUnknownSchemaFields(t *testing.T) {
	cases := map[string]map[string]any{
		"dependency array": {"deps": map[string]string{"future": "format"}},
		"requirement": {"deps": []map[string]any{
			{"name": "future", "req": map[string]string{"version": "^1"}},
		}},
		"release fields": {"name": []string{"serde"}, "vers": 9, "yanked": "false"},
	}
	want := cargo.Release{Name: "serde", Version: "1.0.0", Dependencies: []cargo.Dependency{}}
	for name, fields := range cases {
		t.Run(name, func(t *testing.T) {
			fields["v"] = 3
			unknown := indexLine(t, fields)
			stable := indexLine(t, nil)
			for _, body := range []string{unknown + stable, stable + unknown} {
				got, err := fetchIndex(t, body)
				if err != nil || !reflect.DeepEqual(got, want) {
					t.Fatalf("ожидалось %+v; получено %+v, ошибка %v", want, got, err)
				}
			}
		})
	}
}
