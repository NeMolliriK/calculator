# Распределённый вычислитель арифметических выражений

Этот проект представляет собой распределённую систему для вычисления арифметических выражений. Он принимает выражения через REST API, вычисляет их асинхронно (с использованием распределённого агента) и позволяет получать информацию о статусе и результате вычислений. Помимо API, проект включает простой веб-интерфейс для взаимодействия с системой.

## Особенности

- **Асимметричная архитектура:** Разделение на API-сервер и распределённого агента для выполнения вычислений.
- **Поддержка основных операций:** Сложение, вычитание, умножение, деление (с обработкой ошибок, например, деления на ноль).
- **Асинхронные вычисления:** После отправки запроса вычисление происходит в фоновом режиме.
- **REST API:** Эндпоинты для подачи выражения, получения списка всех вычислений и запроса статуса конкретного выражения.
- **Простой веб-интерфейс:** HTML-страница для ввода выражения и просмотра результатов.
- **Настраиваемость:** Параметры, такие как порт, уровень логирования, время выполнения операций и вычислительная мощность, задаются через файл `.env`.

## Быстрый старт

### Необходимые компоненты

- [Go](https://golang.org/) (версия 1.24+)

### Запуск проекта с использованием Go

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

## Документация API

### 1. Вычисление выражения

**Эндпоинт:** `/api/v1/calculate`  
**Метод:** `POST`  
**Описание:** Принимает JSON с арифметическим выражением для вычисления.

**Тело запроса:**

```json
{
  "expression": "2+2*2"
}
```

**Успешный ответ:**
- **Статус:** `201 Created`
- **Тело ответа:**

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

**Примеры ошибок:**
- **Неподдерживаемый метод (например, GET вместо POST):**

  ```json
  {
    "error": "only POST method is allowed"
  }
  ```

- **Неверный формат JSON:**

  ```json
  {
    "error": "invalid JSON"
  }
  ```

- **Пустое выражение:**

  ```json
  {
    "error": "no expression provided"
  }
  ```

### 2. Получение списка всех вычислений

**Эндпоинт:** `/api/v1/expressions`  
**Метод:** `GET`  
**Описание:** Возвращает список всех отправленных выражений с их статусом и результатом.

**Пример ответа:**
- **Статус:** `200 OK`
- **Тело ответа:**

```json
{
  "expressions": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "status": "completed",
      "result": 6
    },
    {
      "id": "z9y8x7w6-v5u4-3210-tsrq-ponmlkjihgfe",
      "status": "pending",
      "result": 0
    }
  ]
}
```

### 3. Получение результата по ID

**Эндпоинт:** `/api/v1/expressions/{id}`  
**Метод:** `GET`  
**Описание:** Возвращает статус и результат вычисления для выражения с указанным ID.

**Пример успешного ответа:**
- **Статус:** `200 OK`
- **Тело ответа:**

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "completed",
  "result": 6
}
```

**Примеры ошибок:**
- Если ID не передан в URL:

  ```json
  {
    "error": "ID not provided"
  }
  ```

- Если выражение с данным ID отсутствует:

  ```json
  {
    "error": "there is no such expression"
  }
  ```

## Примеры использования cURL

### Успешный запрос на вычисление

```bash
curl --location "http://localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --data "{\"expression\": \"2+2*2\"}"
```

_Ожидаемый ответ:_

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

### Ошибка: Неправильный HTTP-метод

```bash
curl --location "http://localhost:8080/api/v1/calculate" --request GET
```

_Ожидаемый ответ:_

```json
{
  "error": "only POST method is allowed"
}
```

### Ошибка: Пустое выражение

```bash
curl --location "http://localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --data "{\"expression\": \"\"}"
```

_Ожидаемый ответ:_

```json
{
  "error": "no expression provided"
}
```

### Ошибка: Неверный формат JSON

```bash
curl --location "http://localhost:8080/api/v1/calculate" --header "Content-Type: application/json" --data "{invalid json}"
```

_Ожидаемый ответ:_

```json
{
  "error": "invalid JSON"
}
```

## Как это работает

Система организована по модульному принципу:

1. **API-сервер:** Обрабатывает HTTP-запросы от клиентов. Сюда входят эндпоинты для отправки выражения (`/api/v1/calculate`), получения списка вычислений (`/api/v1/expressions`) и получения результата по ID (`/api/v1/expressions/{id}`).
2. **Модуль вычислений:** Реализует алгоритм шантинга (shunting-yard) для преобразования арифметического выражения в обратную польскую запись (RPN) и последующего вычисления результата.
3. **Агент для распределённых вычислений:** Отдельный компонент, который постоянно опрашивает внутренний эндпоинт (`/internal/task`) на предмет новых задач, выполняет вычисления и отправляет результат обратно.
4. **Веб-интерфейс:** Простая HTML-страница (`index.html`), позволяющая пользователю вводить выражения и просматривать результаты вычислений.
5. **Логирование:** Структурированное логирование реализовано для контроля работы сервиса и отладки ошибок.

Ниже представлена упрощённая схема архитектуры:

```
           +----------------------------+
           |       API-сервер           |
           |----------------------------|
           | /api/v1/calculate          |
           | /api/v1/expressions        |
           | /api/v1/expressions/{id}   |
           +-------------+--------------+
                         │
                         │  Отправка задачи
                         ▼
           +----------------------------+
           |  Модуль вычислений         |
           |  (агент для распределённых  |
           |   вычислений)              |
           +-------------+--------------+
                         │
                         │  Обработка задачи
                         ▼
           +----------------------------+
           |     Внутреннее хранилище   |
           |   (in-memory maps: Tasks,  |
           |    Futures)                |
           +----------------------------+
```

## Конфигурация

Основные параметры системы задаются через файл `.env`:

```
PORT=8080
LOG_LEVEL=DEBUG
TIME_ADDITION_MS=1000
TIME_SUBTRACTION_MS=1000
TIME_MULTIPLICATIONS_MS=1000
TIME_DIVISIONS_MS=1000
COMPUTING_POWER=10
```

- **PORT:** Порт, на котором слушает сервер.
- **LOG_LEVEL:** Уровень логирования (DEBUG, INFO, WARN, ERROR).
- **TIME_*_MS:** Симулированное время выполнения для каждой арифметической операции.
- **COMPUTING_POWER:** Количество задач, которые может обрабатывать агент параллельно.

## Тестирование

Чтобы запустить тесты проекта, выполните команду:

```bash
go test ./...
```

Команда запустит все модульные тесты, проверяя корректность работы API, вычислительного модуля и логирования.

---

Наслаждайтесь использованием распределённого вычислителя арифметических выражений!