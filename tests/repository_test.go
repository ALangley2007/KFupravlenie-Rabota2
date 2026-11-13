package tests

import (
	"context"
	"strings"
	"testing"

	"kf2/src/graph"
	"kf2/src/repository"
)

// TestRepositoryWhitespace проверяет пробелы, табуляцию, пустые строки и CRLF.
func TestRepositoryWhitespace(t *testing.T) {
	source := testSource(t, "\r\n A :\t B  C\r\nB:\r\n\r\nC: \r\n")
	got, err := graph.Build(context.Background(), source, graph.Ref{Name: "A"}, "")
	if err != nil || graphText(got) != "A:B,C\nB:\nC:" {
		t.Fatalf("получено %q, ошибка %v", graphText(got), err)
	}
}

// TestRepositoryErrors проверяет формат всех строк, включая недостижимую часть графа.
func TestRepositoryErrors(t *testing.T) {
	cases := []invalidCase{
		{"empty", "\n \n", "пуст"},
		{"no colon", "A B", "строка 1"},
		{"lowercase", "a:", "строка 1"},
		{"long name", "AB:", "строка 1"},
		{"digit", "1:", "строка 1"},
		{"non latin", "А:", "строка 1"},
		{"invalid dependency", "A: b", "зависимости"},
		{"punctuation", "A: B, C", "зависимости"},
		{"second colon", "A: B:\nB:", "зависимости"},
		{"duplicate", "A:\nA:", "строка 2"},
		{"undefined", "A: B", "не объявлена"},
		{"unreachable undefined", "A:\nZ: Y", "не объявлена"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := repository.LoadTest(repositoryFile(t, test.content))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ожидалась ошибка %q, получено %v", test.want, err)
			}
		})
	}
}

// TestRepositoryReadErrors проверяет отсутствие файла и ограничение длины строки.
func TestRepositoryReadErrors(t *testing.T) {
	const longLine = 64 * 1024
	for _, path := range []string{repositoryFile(t, "A:") + ".missing",
		repositoryFile(t, strings.Repeat(" ", longLine))} {
		if _, err := repository.LoadTest(path); err == nil {
			t.Fatal("ожидалась ошибка чтения файла")
		}
	}
}
