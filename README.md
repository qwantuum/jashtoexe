# jashtoexe — Jash → standalone .exe compiler

Превращает `.jash` скрипты в автономные Windows-исполняемые файлы.

## Как это работает

1. Читает `.jash` скрипт
2. Кодирует его в base64
3. Генерирует Go-программу со встроенным интерпретатором Jash + закодированным скриптом
4. Компилирует в `.exe` через `go build`

На выходе — один файл, которому не нужен установленный Jash.

## Использование

```bash
jashtoexe <input.jash> [output.exe]
```

Пример:

```bash
jashtoexe example.jash app.exe
.\app.exe
```

Если `output.exe` не указан — создаётся `output.exe` в текущей папке.

## Сборка jashtoexe

```bash
cd jashtoexe
go build -o jashtoexe.exe .
```

## Зависимости

- Go 1.21+
- Исходный код Jash (лежит рядом, подключается через `replace` в `go.mod`)
- Собираемый `.exe` не требует внешних зависимостей — только Go stdlib

## Структура

```
jashtoexe/
  main.go          # билдер
  jashtoexe.exe    # собранный бинарник
```
