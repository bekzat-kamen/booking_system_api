package main

import (
	"context"
	"log"

	"github.com/bekzat-kamen/booking_system_api/internal/config"
	"github.com/bekzat-kamen/booking_system_api/internal/database"
	"github.com/bekzat-kamen/booking_system_api/internal/handler"
	"github.com/bekzat-kamen/booking_system_api/internal/middleware"
	"github.com/bekzat-kamen/booking_system_api/internal/repository"
	"github.com/bekzat-kamen/booking_system_api/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	log.Printf("Loaded config: APP_ENV=%s, APP_PORT=%s", cfg.AppEnv, cfg.AppPort)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// --- DB ---
	dbConfig := database.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)
	log.Println("Database connection established")

	// --- Redis ---
	redisConfig := database.RedisConfig{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}
	redisClient, err := database.NewRedisConnection(redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}
	defer func() {
		if err := database.CloseRedis(redisClient); err != nil {
			log.Printf("Failed to close redis connection: %v", err)
		}
	}()
	log.Println("Redis connection established")

	// --- JWT ---
	jwtService, err := service.NewJWTService(
		cfg.JWTSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTExpire,
		cfg.JWTRefreshExpire,
	)
	if err != nil {
		log.Fatalf("Failed to create JWT service: %v", err)
	}

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	seatRepo := repository.NewSeatRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	promocodeRepo := repository.NewPromocodeRepository(db)
	dashboardRepo := repository.NewDashboardRepository(db)
	adminUserRepo := repository.NewAdminUserRepository(db)
	adminEventRepo := repository.NewAdminEventRepository(db)
	adminBookingRepo := repository.NewAdminBookingRepository(db)
	adminPromocodeRepo := repository.NewAdminPromocodeRepository(db)

	// --- Services ---
	authService := service.NewAuthService(userRepo, jwtService)
	eventService := service.NewEventService(eventRepo)
	seatService := service.NewSeatService(seatRepo, eventRepo)
	bookingService := service.NewBookingService(bookingRepo, seatRepo, eventRepo)
	paymentService := service.NewPaymentService(paymentRepo, bookingRepo, bookingService)
	promocodeService := service.NewPromocodeService(promocodeRepo)
	dashboardService := service.NewDashboardService(dashboardRepo)
	adminUserService := service.NewAdminUserService(adminUserRepo)
	adminEventService := service.NewAdminEventService(adminEventRepo)
	adminBookingService := service.NewAdminBookingService(adminBookingRepo, seatRepo, eventRepo)
	adminPromocodeService := service.NewAdminPromocodeService(adminPromocodeRepo)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authService)
	eventHandler := handler.NewEventHandler(eventService)
	seatHandler := handler.NewSeatHandler(seatService)
	bookingHandler := handler.NewBookingHandler(bookingService)
	paymentHandler := handler.NewPaymentHandler(paymentService)
	promocodeHandler := handler.NewPromocodeHandler(promocodeService)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	adminUserHandler := handler.NewAdminUserHandler(adminUserService)
	adminEventHandler := handler.NewAdminEventHandler(adminEventService)
	adminBookingHandler := handler.NewAdminBookingHandler(adminBookingService)
	adminPromocodeHandler := handler.NewAdminPromocodeHandler(adminPromocodeService)

	// --- Router ---
	r := gin.Default()

	// Global Rate Limit
	r.Use(middleware.RateLimitMiddleware(redisClient, middleware.DefaultRateLimitConfig))

	api := r.Group("/api/v1")
	{
		// === AUTH (Public + Protected) ===
		auth := api.Group("/auth")
		{
			// Public
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/validate-email", authHandler.ValidateEmail)
			auth.GET("/validate-email", authHandler.ValidateEmail)

			// Protected
			auth.GET("/profile", middleware.AuthMiddleware(jwtService), authHandler.GetProfile)
			auth.PUT("/profile", middleware.AuthMiddleware(jwtService), authHandler.UpdateProfile)
			auth.DELETE("/profile", middleware.AuthMiddleware(jwtService), authHandler.DeactivateProfile)
			auth.POST("/change-password", middleware.AuthMiddleware(jwtService), authHandler.ChangePassword)
		}

		// === EVENTS (Mixed: Public GET, Protected POST/PUT/DELETE) ===
		events := api.Group("/events")
		{
			// Public (No Auth)
			events.GET("", eventHandler.GetAll)
			events.GET("/:id", eventHandler.GetByID)
			events.GET("/:id/seats", seatHandler.GetSeatMap)
			events.GET("/:id/seats/available", seatHandler.GetAvailableSeats)

			// Protected (Auth Required)
			eventsProtected := events.Group("")
			eventsProtected.Use(middleware.AuthMiddleware(jwtService))
			{
				eventsProtected.GET("/organizer", eventHandler.GetByOrganizer)
				eventsProtected.POST("", eventHandler.Create)
				eventsProtected.POST("/:id/publish", eventHandler.PublishEvent)
				eventsProtected.PUT("/:id", eventHandler.Update)
				eventsProtected.DELETE("/:id", eventHandler.Delete)
				eventsProtected.POST("/:id/seats/generate", seatHandler.GenerateSeats)
			}
		}

		// === BOOKINGS (Protected) ===
		bookings := api.Group("/bookings")
		bookings.Use(middleware.AuthMiddleware(jwtService))
		{
			bookings.POST("", bookingHandler.CreateBooking)
			bookings.GET("", bookingHandler.GetUserBookings)
			bookings.GET("/:id", bookingHandler.GetBooking)
			bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
			bookings.POST("/:id/confirm", bookingHandler.ConfirmBooking)
		}

		// === PAYMENTS (Protected) ===
		payments := api.Group("/payments")
		payments.Use(middleware.AuthMiddleware(jwtService))
		{
			payments.POST("", paymentHandler.CreatePayment)
			payments.POST("/:id/process", paymentHandler.ProcessPayment)
			payments.GET("/:id", paymentHandler.GetPayment)
			payments.GET("/booking/:booking_id", paymentHandler.GetPaymentByBooking)
		}
		// Webhook is Public
		api.POST("/payments/webhook", paymentHandler.Webhook)

		// === PROMOCODES (Mixed) ===
		promocodes := api.Group("/promocodes")
		{
			// Public
			promocodes.POST("/validate", promocodeHandler.ValidatePromocode)

			// Protected
			promocodesProtected := promocodes.Group("")
			promocodesProtected.Use(middleware.AuthMiddleware(jwtService))
			{
				promocodesProtected.GET("", promocodeHandler.GetAllPromocodes)
				promocodesProtected.GET("/:id", promocodeHandler.GetPromocode)
				promocodesProtected.POST("", promocodeHandler.CreatePromocode)
				promocodesProtected.PUT("/:id", promocodeHandler.UpdatePromocode)
				promocodesProtected.DELETE("/:id", promocodeHandler.DeletePromocode)
				promocodesProtected.POST("/:id/deactivate", promocodeHandler.DeactivatePromocode)
			}
		}

		// === ADMIN (Admin Only) ===
		admin := api.Group("/admin")
		admin.Use(middleware.AdminMiddleware(jwtService))
		{
			// Dashboard
			admin.GET("/dashboard", dashboardHandler.GetStats)

			// Users Management
			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", adminUserHandler.GetAllUsers)
				adminUsers.GET("/stats", adminUserHandler.GetUserStats)
				adminUsers.GET("/:id", adminUserHandler.GetUserDetail)
				adminUsers.PATCH("/:id/role", adminUserHandler.UpdateUserRole)
				adminUsers.POST("/:id/block", adminUserHandler.BlockUser)
				adminUsers.POST("/:id/unblock", adminUserHandler.UnblockUser)
			}

			// Events Management
			adminEvents := admin.Group("/events")
			{
				adminEvents.GET("", adminEventHandler.GetAllEvents)
				adminEvents.GET("/stats", adminEventHandler.GetEventsStats)
				adminEvents.GET("/:id", adminEventHandler.GetEventDetail)
				adminEvents.PUT("/:id", adminEventHandler.UpdateEvent)
				adminEvents.DELETE("/:id", adminEventHandler.DeleteEvent)
				adminEvents.POST("/:id/publish", adminEventHandler.PublishEvent)
			}

			// Bookings Management
			adminBookings := admin.Group("/bookings")
			{
				adminBookings.GET("", adminBookingHandler.GetAllBookings)
				adminBookings.GET("/stats", adminBookingHandler.GetBookingsStats)
				adminBookings.GET("/export", adminBookingHandler.ExportBookings)
				adminBookings.GET("/:id", adminBookingHandler.GetBookingDetail)
				adminBookings.POST("/:id/cancel", adminBookingHandler.CancelBooking)
				adminBookings.POST("/:id/refund", adminBookingHandler.RefundBooking)
			}

			// Promocodes Management
			adminPromocodes := admin.Group("/promocodes")
			{
				adminPromocodes.GET("", adminPromocodeHandler.GetAllPromocodes)
				adminPromocodes.GET("/stats", adminPromocodeHandler.GetPromocodesStats)
				adminPromocodes.GET("/export", adminPromocodeHandler.ExportPromocodes)
				adminPromocodes.GET("/:id", adminPromocodeHandler.GetPromocodeDetail)
				adminPromocodes.PUT("/:id", adminPromocodeHandler.UpdatePromocode)
				adminPromocodes.DELETE("/:id", adminPromocodeHandler.DeletePromocode)
				adminPromocodes.POST("/bulk-deactivate", adminPromocodeHandler.BulkDeactivate)
			}
		}
	}

	// === Health Check (Public) ===
	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(500, gin.H{"status": "unhealthy", "message": "Database connection failed"})
			return
		}
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			c.JSON(500, gin.H{"status": "unhealthy", "message": "Redis connection failed"})
			return
		}
		c.JSON(200, gin.H{
			"status":   "healthy",
			"database": "connected",
			"redis":    "connected",
		})
	})

	// === Start Server ===
	addr := ":" + cfg.AppPort
	log.Printf("Server starting on http://localhost%s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
