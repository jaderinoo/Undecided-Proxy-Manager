package main

import (
	"log"
	"net/http"
	"os"

	"upm-backend/docs"
	"upm-backend/internal/config"
	"upm-backend/internal/handlers"
	"upm-backend/internal/middleware"
	"upm-backend/internal/services"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           UPM (Undecided Proxy Manager) API
// @version         1.0
// @description     A REST API for managing proxy configurations and users
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:6080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database service
	dbService, err := services.NewDatabaseService()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbService.Close()
	handlers.SetDatabaseService(dbService)
	log.Printf("Database service initialized")

	// Initialize nginx service
	nginxConfigPath := os.Getenv("NGINX_CONFIG_PATH")
	nginxReloadCmd := os.Getenv("NGINX_RELOAD_CMD")
	
	var nginxService *services.NginxService
	if nginxConfigPath != "" && nginxReloadCmd != "" {
		nginxService = services.NewNginxService(nginxConfigPath, nginxReloadCmd)
		handlers.SetNginxService(nginxService)
		log.Printf("Nginx service initialized with config path: %s", nginxConfigPath)
	} else {
		log.Printf("Nginx service not initialized - missing environment variables")
	}

	// Initialize Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "UPM API is running",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Authentication endpoints (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/register", handlers.Register)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Proxy management endpoints
			proxies := protected.Group("/proxies")
			{
				proxies.GET("", handlers.GetProxies)
				proxies.POST("", handlers.CreateProxy)
				proxies.GET("/:id", handlers.GetProxy)
				proxies.PUT("/:id", handlers.UpdateProxy)
				proxies.DELETE("/:id", handlers.DeleteProxy)
			}

			// User management endpoints (admin only)
			users := protected.Group("/users")
			{
				users.GET("", handlers.GetUsers)
				users.POST("", handlers.CreateUser)
				users.GET("/:id", handlers.GetUser)
				users.PUT("/:id", handlers.UpdateUser)
				users.DELETE("/:id", handlers.DeleteUser)
			}
		}
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Initialize docs
	docs.SwaggerInfo.Host = "localhost:" + cfg.BackendPort
	docs.SwaggerInfo.BasePath = "/api/v1"

	log.Printf("Starting UPM API server on port %s", cfg.BackendPort)
	log.Fatal(http.ListenAndServe(":"+cfg.BackendPort, r))
}