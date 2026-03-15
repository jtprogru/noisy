# Noisy — Проект контекста

## Обзор проекта

**Noisy** — это утилита на Go для генерации случайного HTTP/DNS трафика в фоновом режиме. Основная цель — создание "шума" в веб-трафике, чтобы сделать данные о реальном пользовательском трафике менее ценными для сбора и продажи.

### Технологии
- **Язык:** Go 1.21+
- **Основные зависимости:** стандартная библиотека Go (net/http, regexp, encoding/json, compress/gzip)
- **Инструменты разработки:** `go fmt`, `go vet`
- **Контейнеризация:** Docker, docker-compose (multi-stage build)

### Архитектура
Проект состоит из одного основного файла `noisy.go` (~420 строк), который реализует:

**Структуры:**
- `Config` — конфигурация из JSON-файла
- `TimeoutValue` — кастомный тип для обработки timeout (int или bool)
- `Crawler` — основной класс для обхода URL
- `CrawlerTimedOut` — ошибка превышения таймаута

**Основные методы:**
- `LoadConfigFile()` — загрузка конфигурации из JSON
- `request()` — HTTP запрос со случайным User-Agent
- `normalizeLink()` — нормализация ссылок (относительные → абсолютные)
- `isValidURL()` — валидация URL через `url.Parse`
- `extractUrls()` — извлечение ссылок из HTML через regexp
- `browseFromLinks()` — рекурсивный обход ссылок
- `crawl()` — основной цикл обхода root URLs

**Принцип работы:**
1. Загружает конфигурацию из JSON-файла
2. Выбирает случайный URL из списка "root URLs"
3. Извлекает ссылки из HTML-страницы
4. Переходит по случайным ссылкам (с ограничением глубины)
5. Использует случайные User-Agent для каждого запроса
6. Поддерживает настройку timeout для автоматической остановки

## Структура проекта

```
noisy/
├── noisy.go              # Основной файл (класс Crawler)
├── go.mod                # Go модуль
├── config.json           # Конфигурационный файл
├── Makefile              # Make-команды для разработки
├── Dockerfile            # Docker-образ (multi-stage build)
├── docker-compose/
│   └── docker-compose.yml  # Запуск нескольких контейнеров
└── systemd/
    └── noisy.service       # systemd unit для автозапуска
```

## Сборка и запуск

### Локальная сборка

```bash
# Сборка бинарного файла
make build

# Или напрямую
go build -o noisy .

# Запуск
./noisy --config config.json

# Или через make
make run
```

### Docker

```bash
# Сборка образа (multi-stage build)
docker build -t noisy .

# Запуск контейнера
docker run -it noisy --config config.json

# Или через docker-compose (несколько контейнеров)
cd docker-compose
docker-compose up --scale noisy=3
```

### Команды Makefile

| Команда | Описание |
|---------|----------|
| `make build` | Сборка Go бинарника |
| `make run` | Запуск краулера |
| `make fmt` | Форматирование кода (go fmt) |
| `make vet` | Проверка кода (go vet) |
| `make lint` | Запуск всех линтеров (fmt, vet) |
| `make clean` | Очистка бинарных файлов и кэша |
| `make docker.build` | Сборка Docker-образа через docker-compose |
| `make docker.run` | Сборка и запуск контейнера |
| `make docker.logs` | Показать последние 100 строк логов |
| `make docker.logf` | Показать логи в реальном времени |

### Аргументы командной строки

```bash
./noisy --help

# Опции:
# --log          Уровень логирования (debug, info, warning, error), по умолчанию "info"
# --config       Путь к конфигурационному файлу (обязательно)
# --timeout      Время работы в секундах (по умолчанию 0 = без ограничения)
```

## Конфигурация

Файл `config.json` содержит следующие параметры:

```json
{
  "max_depth": 25,           // Максимальная глубина перехода по ссылкам
  "min_sleep": 1,            // Минимальная задержка между запросами (сек)
  "max_sleep": 5,            // Максимальная задержка между запросами (сек)
  "timeout": false,          // Таймаут работы (false/true = без ограничения, или число в секундах)
  "root_urls": [...],        // Список начальных URL для посещения
  "blacklisted_urls": [...], // Список заблокированных URL/паттернов
  "user_agents": [...]       // Список User-Agent для ротации
}
```

**Примечание:** Поле `timeout` может принимать значения:
- `false` — без ограничения по времени (по умолчанию)
- `true` — без ограничения (эквивалентно false)
- число (например, `300`) — остановить работу через N секунд

## Практики разработки

### Код-стиль
- **Форматирование:** go fmt
- **Статический анализ:** go vet

### Запуск линтеров
```bash
make lint
```

### Зависимости
Проект использует только стандартную библиотеку Go:
- `net/http` — HTTP запросы
- `regexp` — регулярные выражения для парсинга ссылок
- `encoding/json` — загрузка конфигурации
- `flag` — аргументы командной строки
- `log` — логирование
- `compress/gzip` — распаковка gzip-ответов
- `io` — чтение тела ответа

## systemd (автозапуск)

Для запуска Noisy как системного сервиса:

```bash
sudo cp systemd/noisy.service /etc/systemd/system
sudo systemctl daemon-reload
sudo systemctl enable noisy && sudo systemctl start noisy
```

Просмотр логов:
```bash
journalctl -f -u noisy
```

## Лицензия

GNU GPLv3 License — см. файл [LICENSE](LICENSE).

## Авторы

- **Itay Hury** ([1tayH](https://github.com/1tayH)) — первоначальная работа на Python
- **Michael Savin** ([jtprogru](https://github.com/jtprogru)) — порт на Go и кастомизация
