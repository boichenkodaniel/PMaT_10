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
