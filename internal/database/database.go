package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type DB struct {
	Postgres *sql.DB
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewPostgresConnection(cfg DBConfig) (*sql.DB, error) {
	// Формируем DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
	)

	// Открываем подключение
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем подключение (пинг)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Настраиваем пул подключений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("✅ PostgreSQL connected successfully")
	return db, nil
}

// Close закрывает подключение
func Close(db *sql.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}
