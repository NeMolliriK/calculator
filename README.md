# Сервис распределённого вычисления арифметических выражений

## Описание

Этот проект реализует HTTP API для асинхронного расчёта арифметических выражений.
Пользователь отправляет выражение на сервер, который ставит его в очередь вычислений, а затем пользователь может запрашивать результат по идентификатору.

## Особенности

- Поддержка базовых арифметических операций: `+`, `-`, `*`, `/`.
- Используется алгоритм обратной польской нотации (RPN) для вычислений.
- Каждая арифметическая операция выполняется с задержкой, имитируя распределённые вычисления.
- Поддержка параллельного выполнения операций.
- Встроенные механизмы логирования запросов и ошибок.
- Веб-интерфейс для отправки и отслеживания выражений.

## API

### Добавление выражения
```bash
curl --location "http://localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --data "{\"expression\": \"2+2*2\"}"
```
Ответ:
```json
{"id": "123e4567-e89b-12d3-a456-426614174000"}
```

### Ошибка 422 (Некорректное выражение):
```bash
curl --location "http://localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --data "{\"expression\": \"2+2a\"}"
```
Ответ:
```json
{"error": "an invalid character is present in the expression: a"}
```

### Ошибка 500 (Внутренняя ошибка сервера):
```bash
curl --location "http://localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --data "{\"expression\": \" \"}"
```
Ответ:
```json
{"error": "internal server error"}
```

### Получение всех выражений
```bash
curl --location 'http://localhost:8080/api/v1/expressions'
```
Ответ:
```json
{"expressions": [
   {"id": "abc123", "status": "processing", "result": 0},
   {"id": "xyz789", "status": "completed", "result": 6}
]}
```

### Получение результата по ID
```bash
curl --location 'http://localhost:8080/api/v1/expressions/abc123'
```
Ответ:
```json
{"id": "abc123", "status": "completed", "result": 4}
```

## Инструкция по запуску

1. Убедитесь, что у вас установлен Go (версии 1.24 или выше).
2. Клонируйте репозиторий:
```bash
git clone https://github.com/NeMolliriK/calculator
cd calculator
```
3. Установите зависимости:
```bash
go mod download github.com/joho/godotenv
go mod download github.com/google/uuid
```
4. Запустите сервис командой:
```bash
go run cmd/main.go
```
5. По умолчанию сервис будет доступен на `http://localhost:8080`.

## Тестирование

Для запуска тестов выполните:
```bash
go test ./...
```

## Архитектура

Проект состоит из нескольких модулей:
- `main.go` – точка входа в приложение.
- `server.go` – реализация HTTP-сервера.
- `handler.go` – обработка HTTP-запросов.
- `calculator.go` – модуль вычислений.
- `application.go` – управление жизненным циклом приложения.
- `loggers.go` – логирование.
- `structures.go` – структуры данных.
- `index.html` – веб-интерфейс.

## Логирование

Логи сохраняются в файлы `server_logs.txt`, `calculations_logs.txt`, `general_logs.txt`.
Пример логов:
```
time=2025-02-27T02:06:43.966+05:00 level=DEBUG msg=tokenize tokens="[{typ:0 val:5} {typ:1 val:*} {typ:2 val:(} {typ:0 val:47.5} {typ:1 val:/} {typ:2 val:(} {typ:0 val:3} {typ:1 val:-} {typ:0 val:7} {typ:3 val:)} {typ:3 val:)} {typ:3 val:)}]"
time=2025-02-28T02:23:25.948+05:00 level=ERROR msg="context canceled"
time=2025-02-26T03:01:33.789+05:00 level=INFO msg="HTTP response" status=200 body="{\"result\":\"-59.375\"}\n" duration=0s
```

## Переменные окружения

- `PORT` – порт сервиса (по умолчанию `8080`).
- `LOG_LEVEL` – уровень логирования (`DEBUG`, `INFO`, `WARN`, `ERROR`).
- `TIME_ADDITION_MS`, `TIME_SUBTRACTION_MS`, `TIME_MULTIPLICATIONS_MS`, `TIME_DIVISIONS_MS` – время выполнения операций (мс).
- `COMPUTING_POWER` – количество горутин, которые одновременно может запустить агент (по умолчанию 10).