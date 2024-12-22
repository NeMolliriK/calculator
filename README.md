# Сервис расчёта арифметических выражений

## Описание

Этот проект предоставляет HTTP API для расчёта арифметических выражений. Пользователь может отправить POST-запрос на единственную точку входа `/api/v1/calculate` с выражением, и получить результат вычислений или ошибку в случае некорректного ввода.

## Особенности

- Поддержка базовых арифметических операций: сложение, вычитание, умножение, деление.
- Обработка выражений с приоритетами операций и скобками.
- Обработка ошибок, включая некорректный ввод и деление на ноль.
- Встроенные механизмы логирования запросов и ошибок.
- Простое развертывание и запуск.

## Использование

Пример запроса через `curl`:

### Успешный результат:
```bash
curl --location "http://localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --data "{\"expression\": \"2+2*2\"}"
```
Ответ:
```json
{"result": "6"}
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

## Инструкция по запуску

1. Убедитесь, что у вас установлен Go (версии 1.23.2 или выше).
2. Клонируйте репозиторий:
```bash
git clone https://github.com/NeMolliriK/calculator
cd calculator
```
3. Инициализируйте Go-модуль (если этого ещё не сделано):
```bash
go mod init calculator
```
4. Установите зависимости:
```bash
go mod download github.com/joho/godotenv
```
5. Запустите сервис командой:
```bash
go run cmd/main.go
```
6. По умолчанию сервис будет доступен на `http://localhost:8080`.

## Тестирование

Для запуска тестов выполните:
```bash
go test ./...
```

## Работа с .env

Для настройки параметров проекта можно использовать файл `.env`. Пример содержимого:

```env
PORT=8080
LOG_LEVEL=INFO
```

- **PORT**: порт, на котором запускается сервис. Если не указан, используется значение по умолчанию `8080`.
- **LOG_LEVEL**: уровень логирования. Поддерживаются значения `DEBUG`, `INFO`, `WARN`, `ERROR`. Значение по умолчанию — `INFO`.

## Архитектура

Проект разделён на несколько модулей:

- **calculator**: содержит логику для вычисления выражений.
- **server**: реализует HTTP-сервер и обработку запросов.
- **application**: точка входа и настройка приложения.

## Пример логирования

Пример записи в лог-файл:
```
time=2024-12-23T01:05:17.905+05:00 level=INFO msg="HTTP Method: POST, Request Body: {    expression: 5+5}"
time=2024-12-23T01:05:17.911+05:00 level=INFO msg="HTTP Response Code: 200, Response Body: {result:10}"
```

## Расширение функциональности

Возможные улучшения проекта:

1. **Поддержка дополнительных операторов:** Добавление операций, таких как возведение в степень или модуль.
2. **Обработка пользовательских переменных:** Возможность использовать переменные в выражениях (например, `x=2; x+2`).
3. **Интернационализация:** Перевод сообщений об ошибках и интерфейса на несколько языков.
4. **Веб-интерфейс:** Реализация графического пользовательского интерфейса для ввода выражений и просмотра результатов.
5. **Поддержка истории запросов:** Хранение выполненных запросов для их последующего анализа или повторного использования.

## Поддержка и обратная связь

Если у вас возникли вопросы или предложения по улучшению, пожалуйста, создайте issue в репозитории проекта или свяжитесь с разработчиками через контактные данные, указанные в профиле GitHub.
