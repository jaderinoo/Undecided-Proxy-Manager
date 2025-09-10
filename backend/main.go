package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"upm-backend/docs"
	"upm-backend/internal/auth"
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

// ensureAdminUserExists manages admin user based on ADMIN_PASSWORD environment variable
func ensureAdminUserExists(dbService *services.DatabaseService, cfg *config.Config) error {
	// Check if admin user exists in database
	exists, err := dbService.AdminUserExists()
	if err != nil {
		return fmt.Errorf("failed to check admin user existence: %w", err)
	}

	// If ADMIN_PASSWORD is not set, ensure no admin user exists
	if cfg.AdminPassword == "" {
		if exists {
			log.Println("ADMIN_PASSWORD not set, removing admin user...")
			if err := dbService.DeleteAdminUser(); err != nil {
				return fmt.Errorf("failed to delete admin user: %w", err)
			}
		} else {
			log.Println("No admin user exists and ADMIN_PASSWORD not set - admin access disabled")
		}
		return nil
	}

	// ADMIN_PASSWORD is set, ensure admin user exists and password is correct
	if !exists {
		// Create admin user
		log.Println("ADMIN_PASSWORD set but no admin user exists, creating one...")
		hashedPassword, err := auth.HashPassword(cfg.AdminPassword)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		if err := dbService.CreateAdminUser(hashedPassword); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		log.Println("Admin user created successfully")
	} else {
		// Admin user exists, check if password needs updating
		adminUser, err := dbService.GetAdminUser()
		if err != nil {
			return fmt.Errorf("failed to get admin user: %w", err)
		}

		// Check if the stored password matches the current ADMIN_PASSWORD
		if !auth.CheckPasswordHash(cfg.AdminPassword, adminUser.Password) {
			log.Println("ADMIN_PASSWORD changed, updating admin user password...")
			hashedPassword, err := auth.HashPassword(cfg.AdminPassword)
			if err != nil {
				return fmt.Errorf("failed to hash admin password: %w", err)
			}

			if err := dbService.UpdateAdminUserPassword(hashedPassword); err != nil {
				return fmt.Errorf("failed to update admin user password: %w", err)
			}
		} else {
			log.Println("Admin user exists with correct password")
		}
	}

	return nil
}

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

	// Ensure admin user exists
	if err := ensureAdminUserExists(dbService, cfg); err != nil {
		log.Fatalf("Failed to ensure admin user exists: %v", err)
	}

	// Initialize nginx service
	nginxConfigPath := os.Getenv("NGINX_CONFIG_PATH")
	nginxReloadCmd := os.Getenv("NGINX_RELOAD_CMD")
	nginxContainerName := os.Getenv("NGINX_CONTAINER_NAME")

	var nginxService *services.NginxService
	if nginxConfigPath != "" && nginxReloadCmd != "" {
		nginxService = services.NewNginxService(nginxConfigPath, nginxReloadCmd, nginxContainerName, dbService)
		handlers.SetNginxService(nginxService)
		log.Printf("Nginx service initialized with config path: %s, container: %s", nginxConfigPath, nginxContainerName)
	} else {
		log.Printf("Nginx service not initialized - missing environment variables")
	}

	// Initialize Docker service
	dockerService, err := services.NewDockerService()
	if err != nil {
		log.Printf("Docker service not initialized: %v", err)
	} else {
		handlers.SetDockerService(dockerService)
		log.Printf("Docker service initialized")
	}

	// Initialize DNS service
	dnsService := services.NewDNSService(dbService)
	handlers.SetDNSService(dnsService)
	log.Printf("DNS service initialized")

	// Initialize scheduler service
	schedulerService := services.NewSchedulerService(dnsService)
	handlers.SetSchedulerService(schedulerService)
	log.Printf("Scheduler service initialized")

	// Load and start scheduled DNS update jobs
	if err := schedulerService.LoadAndStartJobs(); err != nil {
		log.Printf("Warning: Failed to load scheduled jobs: %v", err)
	}

	// Start periodic DNS updates
	dnsService.StartPeriodicUpdates()

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
				proxies.GET("/:id/certificate", handlers.GetProxyCertificate)
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

			// Container management endpoints
			containers := protected.Group("/containers")
			{
				containers.GET("", handlers.GetContainers)
				containers.GET("/:id", handlers.GetContainer)
				containers.GET("/:id/stats", handlers.GetContainerStats)
			}

			// Nginx management endpoints
			nginx := protected.Group("/nginx")
			{
				nginx.POST("/reload", handlers.ReloadNginx)
				nginx.POST("/test", handlers.TestNginxConfig)
				nginx.POST("/regenerate-config", handlers.RegenerateProxyConfig)
				nginx.GET("/admin-ip-restrictions", handlers.GetAdminIPRestrictions)
				nginx.PUT("/admin-ip-restrictions", handlers.UpdateAdminIPRestrictions)
			}

			// DNS management endpoints
			dns := protected.Group("/dns")
			{
				// DNS Config endpoints
				dns.GET("/configs", handlers.GetDNSConfigs)
				dns.POST("/configs", handlers.CreateDNSConfig)
				dns.GET("/configs/:id", handlers.GetDNSConfig)
				dns.PUT("/configs/:id", handlers.UpdateDNSConfig)
				dns.DELETE("/configs/:id", handlers.DeleteDNSConfig)

				// DNS Record endpoints
				dns.GET("/records", handlers.GetDNSRecords)
				dns.POST("/records", handlers.CreateDNSRecord)
				dns.GET("/records/:id", handlers.GetDNSRecord)
				dns.PUT("/records/:id", handlers.UpdateDNSRecord)
				dns.DELETE("/records/:id", handlers.DeleteDNSRecord)

				// DNS Update endpoints
				dns.POST("/records/:id/update", handlers.UpdateDNSRecordNow)
				dns.POST("/update-all", handlers.UpdateAllDNSRecords)
				dns.GET("/status", handlers.GetDNSStatus)
				dns.GET("/public-ip", handlers.GetPublicIP)
				dns.GET("/scheduled-jobs", handlers.GetScheduledJobs)
				dns.POST("/scheduled-jobs/:recordId/pause", handlers.PauseScheduledJob)
				dns.POST("/scheduled-jobs/:recordId/resume", handlers.ResumeScheduledJob)
			}

			// Certificate management endpoints
			certificates := protected.Group("/certificates")
			{
				certificates.GET("", handlers.GetCertificates)
				certificates.POST("", handlers.CreateCertificate)
				certificates.POST("/letsencrypt", handlers.GenerateLetsEncryptCertificate)
				certificates.GET("/:id", handlers.GetCertificate)
				certificates.PUT("/:id", handlers.UpdateCertificate)
				certificates.DELETE("/:id", handlers.DeleteCertificate)
				certificates.GET("/:id/proxies", handlers.GetCertificateProxies)
				certificates.POST("/:id/renew", handlers.RenewCertificate)
			}

			// Settings management endpoints
			settings := protected.Group("/settings")
			{
				settings.GET("", handlers.GetSettings)
				settings.PUT("", handlers.UpdateSettings)
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
