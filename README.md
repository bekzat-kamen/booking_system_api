![Go CI](https://github.com/bekzat-kamen/booking_system_api/actions/workflows/go.yml/badge.svg)
[![codecov](https://codecov.io/github/bekzat-kamen/booking_system_api/branch/main/graph/badge.svg)](https://codecov.io/github/bekzat-kamen/booking_system_api)

# 🎫 Booking System API

Современная и высокопроизводительная система управления мероприятиями и бронирования билетов, написанная на языке **Go**. Проект спроектирован с учетом масштабируемости, безопасности и удобства использования.

---

## 🚀 Основной функционал

- 🔐 **Аутентификация**: Полноценная работа с JWT (Access & Refresh токены), смена пароля, подтверждение email.
- 🎫 **Мероприятия**: CRUD операции, система статусов (черновик, опубликован, отменен), поиск и фильтрация.
- 🪑 **Места**: Автоматическая генерация схемы зала, бронирование конкретных рядов и мест.
- 📋 **Бронирования**: Механизм резервирования с автоматическим освобождением мест при неоплате (TTL).
- 💳 **Платежи**: Имитация платежного шлюза, обработка вебхуков, статистика доходов.
- 🏷️ **Промокоды**: Гибкая система скидок с проверкой лимитов использования и сроков действия.
- 🛡️ **Безопасность**: Rate limiting (ограничение запросов) через Redis, защита маршрутов по ролям.
- 📊 **Админ-панель**: Комплексный дашборд со статистикой в реальном времени, управление пользователями и логами.
- 📝 **Логирование**: Структурированные JSON-логи (slog) для удобного мониторинга.

---

## 🛠 Технологический стек

- **Язык**: Go 1.25+
- **HTTP Framework**: [Gin Gonic](https://github.com/gin-gonic/gin)
- **Database**: PostgreSQL 16 (с использованием [sqlx](https://github.com/jmoiron/sqlx))
- **Cache / Rate Limit**: Redis 7
- **Documentation**: [Swagger / OpenAPI 3.0](https://github.com/swaggo/swag)
- **Containerization**: Docker & Docker Compose
- **Testing**: Testify, SQLMock, RedisMock

---

## 📦 Быстрый запуск (Docker)

Самый простой способ запустить проект — использовать Docker Compose. Это поднимет API, PostgreSQL и Redis одной командой.

1. **Клонируйте репозиторий:**
   ```bash
   git clone https://github.com/bekzat-kamen/booking_system_api.git
   cd booking_system_api
   ```

2. **Запустите контейнеры:**
   ```bash
   docker-compose up -d --build
   ```

3. **Готово!** API будет доступно по адресу: `http://localhost:8080`

---

## 📖 Документация API

После запуска сервера интерактивная документация Swagger доступна по адресу:
👉 **[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

Здесь вы можете протестировать все доступные эндпоинты прямо из браузера.

---

## 🧪 Тестирование и покрытие

Проект покрыт unit-тестами более чем на **60%**. Мы используем моки для базы данных и внешних сервисов.

**Запуск всех тестов:**
```bash
go test ./...
```

**Просмотр покрытия кода:**
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## 📂 Структура проекта

```text
├── cmd/api/            # Точка входа в приложение
├── docs/               # Сгенерированная документация Swagger
├── internal/
│   ├── config/         # Загрузка конфигурации (.env)
│   ├── database/       # Подключение к DB и Redis
│   ├── handler/        # HTTP хендлеры (контроллеры)
│   ├── middleware/     # Промежуточное ПО (Auth, Logger, RateLimit)
│   ├── model/          # Структуры данных и БД модели
│   ├── repository/     # Уровень работы с БД (SQL запросы)
│   └── service/        # Бизнес-логика приложения
├── migrations/         # SQL файлы миграций
└── docker-compose.yml  # Оркестрация контейнеров
```

---

## ⚙️ Конфигурация

Все настройки хранятся в файле `.env`. Пример настроек можно найти в `.env.example`.

| Переменная | Описание | Дефолт |
|------------|----------|--------|
| `APP_PORT` | Порт приложения | `8080` |
| `DB_HOST`  | Хост базы данных | `db` (в Docker) |
| `REDIS_HOST`| Хост Redis | `redis` (в Docker) |
| `JWT_SECRET`| Секрет для токенов| `your-secret-key` |

---
