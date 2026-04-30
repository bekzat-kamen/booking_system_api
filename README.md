![Go CI](https://github.com/bekzat-kamen/booking_system_api/actions/workflows/go.yml/badge.svg)

# 🎫 Booking System API

Система управления мероприятиями и бронирования билетов на Go.

---

## 🚀 Tech Stack

- **Go** 1.26+
- **Gin** — HTTP фреймворк
- **PostgreSQL** 15 — база данных
- **Redis** 7 — кэш и rate limiting
- **JWT** — аутентификация
- **sqlx** — работа с БД
- **golang-migrate** — миграции

---

## ✅ Реализованный функционал

- 🔐 Регистрация, вход, JWT токены (Access + Refresh)
- 🎫 CRUD мероприятий с модерацией
- 🪑 Генерация мест и схема зала
- 📋 Бронирования с таймаутом оплаты
- 💳 Платежи (имитация + webhook)
- 🏷️ Промокоды со скидками
- 🛡️ Rate Limiting через Redis
- 📊 Админ-панель (дашборд, пользователи)

---
