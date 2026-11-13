#!/bin/sh
set -eu

project_dir=$(CDPATH= cd "$(dirname "$0")" && pwd)
cd "$project_dir"

usage() {
    cat <<'EOF'
Использование: ./run.sh [команда] [аргументы]

Команды:
  run [аргументы]     Запустить приложение (команда по умолчанию)
  build [флаги]       Собрать приложение в bin/kf2
  test [флаги]        Выполнить все тесты
  check [флаги]       Выполнить go vet
  demo                Продемонстрировать графы, порядок загрузки и ошибки
  compare             Сравнить порядок с Cargo (нужны Cargo и rustc)
  help                Показать эту справку

Примеры:
  ./run.sh
  ./run.sh run -config config.xml
  ./run.sh run -h
  ./run.sh test -race -count=1
EOF
}

require_no_arguments() {
    if [ "$#" -ne 0 ]; then
        printf 'Ошибка: команда не принимает аргументы\n' >&2
        exit 1
    fi
}

make_config() {
    printf '%s\n' '<config>' \
        "<package>$1</package>" \
        "<repository>$2</repository>" \
        "<test_mode>$3</test_mode>" \
        "<filter>$4</filter>" '</config>' > "$temp_dir/invalid.xml"
}

expect_error() {
    printf '\nНекорректный параметр %s:\n' "$1"
    status=0
    "$temp_dir/kf2" -config "$temp_dir/invalid.xml" \
        > "$temp_dir/stdout" 2> "$temp_dir/stderr" || status=$?
    if [ "$status" -ne 1 ] || [ -s "$temp_dir/stdout" ]; then
        printf 'Неожиданный код завершения или непустой stdout\n' >&2
        exit 1
    fi
    error_text=$(cat "$temp_dir/stderr")
    case "$error_text" in
        *"Ошибка: $1:"*) printf '%s\n' "$error_text" ;;
        *) printf 'Неожиданная ошибка: %s\n' "$error_text" >&2; exit 1 ;;
    esac
}

demo() {
    temp_dir=$(mktemp -d)
    trap 'rm -rf "$temp_dir"' EXIT
    trap 'exit 1' HUP INT TERM
    go build -o "$temp_dir/kf2" ./src

    printf 'Тестовые графы: BFS, транзитивность, фильтрация и циклы:\n'
    go test -count=1 -v -run '^TestStage3Demo$' ./tests

    printf '\nПорядок загрузки: зависимости перед пакетом, фильтры и ошибки циклов:\n'
    go test -count=1 -v -run '^TestStage4Demo$' ./tests

    make_config 'bad/name' 'https://index.crates.io' 'false' ''
    expect_error package
    make_config 'serde' 'repository.txt' 'false' ''
    expect_error repository
    make_config 'serde' 'https://index.crates.io' 'yes' ''
    expect_error test_mode
    make_config 'serde' 'https://index.crates.io' 'false' '&#10;'
    expect_error filter
    make_config 'serde' 'https://index.crates.io' 'false' ''
    sed 's|</config>|<load_order>yes</load_order></config>|' "$temp_dir/invalid.xml" > "$temp_dir/order.xml"
    mv "$temp_dir/order.xml" "$temp_dir/invalid.xml"
    expect_error load_order
}

compare() {
    if ! command -v cargo >/dev/null 2>&1 || ! command -v rustc >/dev/null 2>&1; then
        printf 'Ошибка: для сравнения нужны Cargo и rustc в PATH\n' >&2
        exit 1
    fi
    KF2_CARGO_COMPARE=true go test -count=1 -v -run '^TestCargoComparison$' ./tests
}

command=${1:-run}
if [ "$#" -gt 0 ]; then
    shift
fi

case "$command" in
    run) exec go run ./src "$@" ;;
    build)
        mkdir -p bin
        exec go build -o bin/kf2 "$@" ./src
        ;;
    test) exec go test "$@" ./... ;;
    check) exec go vet "$@" ./... ;;
    demo)
        require_no_arguments "$@"
        demo
        ;;
    compare)
        require_no_arguments "$@"
        compare
        ;;
    help|-h|--help)
        require_no_arguments "$@"
        usage
        ;;
    -*) exec go run ./src "$command" "$@" ;;
    *)
        printf 'Ошибка: неизвестная команда "%s"\n' "$command" >&2
        usage >&2
        exit 1
        ;;
esac
