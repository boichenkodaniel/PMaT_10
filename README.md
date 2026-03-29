# Лабораторная работа №10

## Дисциплина
Методы и технологии программирования

## Студент
Бойченко Даниэль

## Группа
220032-11

## Вариант
2

## Выполненные задания

### Средняя сложность:

#### 2. Добавить middleware для логирования в Go.
#### 4. Создать FastAPI-сервис, который вызывает Go-сервис через HTTP.
#### 6. Сравнить скорость ответа FastAPI и Gin под нагрузкой (wrk/ab).

### Повышенная сложность:

#### 2. Создать API-шлюз на Go, который маршрутизирует запросы к разным микросервисам (Python, Go).
#### 4. Использовать WebSocket: реализовать чат на Go и подключиться к нему из Python.

---

## Инструкции по запуску

### Task 1 — Middleware для логирования в Go

**Расположение:** `task1/`

**Запуск сервера:**
```bash
cd task1
go run main.go
```
Сервер запустится на порту `8080`.

**Запуск тестов:**
```bash
cd task1
pytest test_main.py -v
```
Тесты автоматически запускают сервер перед проверкой.

**Эндпоинты:**
- `GET /` — возвращает "Hello, World!"
- `GET /health` — возвращает "OK"
- `GET /slow` — имитирует медленный запрос (2 сек)

---

### Task 2 — FastAPI + Go сервисы

**Расположение:** `task2/`

**Структура:**
- `main.py` — FastAPI сервис (порт 8000)
- `go-service/` — Go сервис (порт 8080)

**Запуск Go сервиса:**
```bash
cd task2/go-service
go run main.go
```

**Запуск FastAPI сервиса:**
```bash
cd task2
python main.py
```

**Запуск тестов:**
```bash
cd task2
pytest test_main.py -v
```

**Эндпоинты FastAPI:**
- `GET /health` — проверка здоровья обоих сервисов
- `POST /add` — сложение чисел (вызывает Go сервис)
- `POST /subtract` — вычитание чисел (вызывает Go сервис)

**Эндпоинты Go:**
- `GET /health` — проверка здоровья
- `POST /add` — сложение чисел
- `POST /subtract` — вычитание чисел

---

### Task 3 — Сравнение производительности FastAPI vs Gin

**Расположение:** `task3/`

**Структура:**
- `main.py` — FastAPI сервер (порт 8000)
- `main.go` — Gin сервер (порт 8080)
- `benchmark.py` — скрипт для нагрузочного тестирования
- `test_main.py` — pytest тесты для FastAPI
- `main_test.go` — Go тесты для Gin

**Запуск FastAPI сервера:**
```bash
cd task3
pip install -r requirements.txt
python main.py
```

**Запуск Gin сервера:**
```bash
cd task3
go mod tidy
go run main.go
```

**Запуск тестов FastAPI:**
```bash
cd task3
pytest test_main.py -v
```

**Запуск тестов Gin:**
```bash
cd task3
go test -v ./...
```

**Запуск бенчмарка:**
```bash
cd task3
python benchmark.py
```

**Эндпоинты (оба сервера):**
- `GET /` — приветственное сообщение
- `GET /health` — проверка здоровья
- `POST /echo` — эхо сообщения с длиной
- `GET /items` — получить все элементы
- `POST /items` — создать элемент
- `GET /items/{id}` — получить элемент по ID
- `PUT /items/{id}` — обновить элемент
- `DELETE /items/{id}` — удалить элемент

**Пример запроса:**
```bash
# Health check
curl http://localhost:8000/health
curl http://localhost:8080/health

# Создать элемент
curl -X POST http://localhost:8000/items \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "price": 19.99}'
```

---

### Task 4 — API-шлюз на Go для микросервисов (Python, Go)

**Расположение:** `task4/`

**Структура:**
- `gateway/main.go` — API Gateway (порт 8080)
- `user-service/main.go` — User Service на Go (порт 8081)
- `order-service/app.py` — Order Service на Python/Flask (порт 8082)
- `gateway/main_test.go` — Go тесты для gateway
- `user-service/main_test.go` — Go тесты для user-service
- `test_services.py` — Pytest тесты (order service + integration)

**Установка зависимостей:**
```bash
cd task4
pip install -r requirements.txt
```

**Запуск User Service (Go):**
```bash
cd task4/user-service
go run main.go
```

**Запуск Order Service (Python):**
```bash
cd task4/order-service
python app.py
```

**Запуск API Gateway (Go):**
```bash
cd task4/gateway
go run main.go
```

**Запуск Go тестов:**
```bash
cd task4
go test -v ./...
```

**Запуск Pytest тестов:**
```bash
cd task4
python -m pytest test_services.py -v
```

**Эндпоинты Gateway:**
- `GET /api/user?id={id}` — получить информацию о пользователе
- `GET /api/orders?user_id={id}` — получить заказы пользователя
- `GET /api/profile?id={id}` — получить профиль (пользователь + заказы)
- `GET /health` — проверка здоровья шлюза и сервисов
- `GET /` — информация об API

**Примеры запросов:**
```bash
# Получить пользователя
curl "http://localhost:8080/api/user?id=1"

# Получить заказы
curl "http://localhost:8080/api/orders?user_id=1"

# Получить полный профиль
curl "http://localhost:8080/api/profile?id=1"

# Проверка здоровья
curl "http://localhost:8080/health"
```

---

### Task 5 — WebSocket чат на Go с Python-клиентом

**Расположение:** `task5/`

**Структура:**
- `main.go` — WebSocket сервер на Go (порт 8080)
- `main_test.go` — Go тесты для сервера
- `client.py` — интерактивный Python-клиент
- `test_chat.py` — Pytest тесты для проверки чата

**Установка зависимостей:**
```bash
cd task5
pip install websockets pytest pytest-asyncio
```

**Запуск сервера:**
```bash
cd task5
go mod tidy
go run main.go
```
Сервер запустится на порту `8080`.

**Запуск Go тестов:**
```bash
cd task5
go test -v
```

**Запуск Pytest тестов:**
```bash
cd task5
python -m pytest test_chat.py -v
```

**Запуск Python-клиента:**
```bash
cd task5
python client.py
```

**Демо-режим клиента:**
```bash
cd task5
python client.py --demo
```

**Команды клиента:**
- `/nick <name>` — сменить никнейм
- `/quit` — выйти из чата
- `/help` — показать справку
- `<сообщение>` — отправить сообщение в чат

**Эндпоинты сервера:**
- `GET /ws` — WebSocket подключение
- `GET /clients` — количество подключенных клиентов (JSON)

**Примеры запросов:**
```bash
# Проверка количества клиентов
curl http://localhost:8080/clients

# WebSocket подключение (через wscat или клиент)
wscat -c ws://localhost:8080/ws
```

**Формат сообщений:**
```json
// Отправка сообщения
{"type": "chat", "content": "Hello!"}

// Смена никнейма
{"type": "nick_change", "new_nickname": "NewName"}
```

---

## Требования

### Python
- Python 3.10+
- `pip install -r requirements.txt` (для каждого задания)

### Go
- Go 1.21+
- `go mod tidy` (для каждого задания)

### Тесты
- `pytest` для Python тестов
- Встроенный `testing` для Go тестов
