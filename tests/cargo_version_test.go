package tests

import "testing"

// TestCargoSelectVersion проверяет SemVer независимо от порядка строк в индексе.
func TestCargoSelectVersion(t *testing.T) {
	cases := [][]string{
		{"1.9.9", "1.10.0"}, {"1.99.9", "2.0.0"}, {"1.0.9", "1.0.10"},
		{"1.0.0", "1.0.1+metadata.01"}, {"99.0.0-beta", "1.0.0"},
		{"1.0.0-alpha", "1.0.0-alpha.1"}, {"1.0.0-alpha.1", "1.0.0-alpha.beta"},
		{"1.0.0-alpha.beta", "1.0.0-beta"}, {"1.0.0-beta", "1.0.0-beta.2"},
		{"1.0.0-beta.2", "1.0.0-beta.11"}, {"1.0.0-beta.11", "1.0.0-rc.1"},
		{"1.0.0-rc.1", "1.0.0"}, {"1.0.0-beta", "2.0.0-alpha"},
		{"1.0.0-9", "1.0.0-10"}, {"1.0.0-1", "1.0.0-a"},
		{"999999999999999999999.0.0", "1000000000000000000000.0.0"},
	}
	for _, pair := range cases {
		t.Run(pair[1], func(t *testing.T) {
			older := indexLine(t, map[string]any{"vers": pair[0]})
			newer := indexLine(t, map[string]any{"vers": pair[1]})
			for _, body := range []string{older + newer, newer + older} {
				got, err := fetchIndex(t, body)
				if err != nil || got.Version != pair[1] {
					t.Fatalf("ожидалось %s; получено %+v, ошибка %v", pair[1], got, err)
				}
			}
		})
	}
}

// TestCargoInvalidVersions проверяет некорректные версии, в том числе ведущие нули.
func TestCargoInvalidVersions(t *testing.T) {
	versions := []string{
		"", "1.0", "1.2.3.4", "v1.2.3", "-1.0.0", "01.0.0", "1.00.0", "1.0.01",
		"1.0.0-", "1.0.0+", "1.0.0-01", "1.0.0-rc..1", "1.0.0-rc.01", "1.0.0+meta..data",
		"1.0.0+meta+data", "1.0.0-pre_release", "1.0.0+мета", " 1.0.0", "1.0.0 ",
	}
	for _, value := range versions {
		t.Run(value, func(t *testing.T) {
			if _, err := fetchIndex(t, indexLine(t, map[string]any{"vers": value})); err == nil {
				t.Fatalf("недопустимая версия %q принята", value)
			}
		})
	}
}
