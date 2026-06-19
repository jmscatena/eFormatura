package main

import (
	"log"
	"time"

	"backend/config"
	"backend/internal/handlers"
	"backend/internal/middleware"
	"backend/internal/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	config.ConnectDB()
	models.AutoMigrate(config.DB)

	// Inicializar rate limiter PostgreSQL para autenticação
	middleware.InitPostgresRateLimiter()

	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Helmet security headers — aplicar em TODAS as rotas
	r.Use(middleware.HelmetMiddleware())

	// Configurar CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://formatura.sytes.net", "http://localhost:8000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Auth routes — com rate limiter PostgreSQL (proteger contra brute force)
	authGroup := r.Group("")
	authGroup.Use(middleware.PostgresLoginRateLimiter())
	{
		authGroup.POST("/login", handlers.Login)
		authGroup.POST("/register", handlers.Register)
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthRequired())
	protected.Use(middleware.RateLimit(100, time.Minute)) // 100 requests per minute
	{
		// Everyone can read
		protected.GET("/incomes", handlers.GetIncomes)
		protected.GET("/expenses", handlers.GetExpenses)
		protected.GET("/notifications", handlers.ListNotifications)

		// Admin only: write and users
		adminGroup := protected.Group("")
		adminGroup.Use(middleware.RoleRequired("admin"))
		{
			// Users management
			adminGroup.GET("/users", handlers.GetUsers)
			adminGroup.PUT("/users/:id", handlers.UpdateUser)
			adminGroup.PUT("/users/:id/password", handlers.ResetPassword)
			adminGroup.PUT("/users/:id/disable", handlers.DisableUser)

			// Incomes
			adminGroup.POST("/incomes", handlers.CreateIncome)
			adminGroup.PUT("/incomes/:id", handlers.UpdateIncome)
			adminGroup.DELETE("/incomes/:id", handlers.DeleteIncome)

			// Expenses
			adminGroup.POST("/expenses", handlers.CreateExpense)
			adminGroup.PUT("/expenses/:id", handlers.UpdateExpense)
			adminGroup.DELETE("/expenses/:id", handlers.DeleteExpense)

			// Installments
			adminGroup.POST("/installments", handlers.CreateInstallment)
			adminGroup.PUT("/installments/:id", handlers.UpdateInstallment)
			adminGroup.PUT("/installments/:id/pay", handlers.PayInstallment)
			adminGroup.DELETE("/installments/:id", handlers.DeleteInstallment)

			// Notifications
			adminGroup.POST("/notifications/:id/read", handlers.MarkNotificationAsRead)
			adminGroup.POST("/notifications/read/all", handlers.MarkAllNotificationsAsRead)
		}
	}

	r.Run() // listen and serve on 0.0.0.0:8080
}
