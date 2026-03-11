package main

import (
	"log"

	"github.com/bekzat-kamen/booking_system_api/internal/config"
	"github.com/bekzat-kamen/booking_system_api/internal/database"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	log.Printf("📦 Loaded config: APP_ENV=%s, APP_PORT=%s", cfg.AppEnv, cfg.AppPort)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	dbConfig := database.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}

	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer database.Close(db)
	log.Println("🗄️ Database connection established")

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "🎫 Booking System API",
			"env":     cfg.AppEnv,
			"status":  "running",
		})
	})

	r.GET("/health", func(c *gin.Context) {

		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{
				"status":  "unhealthy",
				"message": "Database connection failed",
			})
			return
		}

		c.JSON(200, gin.H{
			"status":   "healthy",
			"database": "connected",
		})
	})

	addr := ":" + cfg.AppPort
	log.Printf("🚀 Server starting on http://localhost%s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
