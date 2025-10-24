package main

import (
	"log"
	"os"
	_ "treblle_project/docs"
	"treblle_project/internal/database"
	"treblle_project/internal/handlers"
	"treblle_project/internal/jikan"
	"treblle_project/internal/repository"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// @title           Treblle API Monitor
// @version         1.0
// @description     API monitoring service that proxies requests to Jikan API and tracks performance metrics and issues.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  lovro.dvorski@outlook.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api

// @schemes http https

// @tag.name requests
// @tag.description Operations for viewing API request call logs

// @tag.name problems
// @tag.description Operations for viewing detected failed or problematic API calls

// @tag.name jikan
// @tag.description Proxy to Jikan API with monitoring

func main() {
	// Initialize database with configurable path
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./api_monitor.db"
	}

	log.Printf("Using database path: %s", dbPath)
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	requestRepo := repository.NewRequestRepository(db)
	problemRepo := repository.NewProblemRepository(db)

	// Initialize Jikan client
	jikanClient := jikan.NewClient()

	// Initialize handlers
	requestHandler := handlers.NewRequestHandler(requestRepo)
	problemHandler := handlers.NewProblemHandler(problemRepo)
	jikanHandler := handlers.NewJikanHandler(jikanClient, requestRepo, problemRepo)

	// Setup router
	r := gin.Default()

	// API routes
	api := r.Group("/api")
	{
		// Request viewing endpoints
		api.GET("/requests", requestHandler.ListRequests)
		api.GET("/requests/table", requestHandler.TableView)
		api.GET("/requests/csv", requestHandler.CSVExport)

		// Problem viewing endpoints
		api.GET("/problems", problemHandler.ListProblems)
		api.GET("/problems/table", problemHandler.TableView)
		api.GET("/problems/csv", problemHandler.CSVExport)

		// Jikan proxy endpoint - matches any path
		api.GET("/jikan/*path", jikanHandler.ProxyRequest)
	}

	// Health check
	// @Summary      Health check
	// @Description  Check if the API is running
	// @Tags         health
	// @Produce      json
	// @Success      200  {object}  map[string]string
	// @Router       /health [get]
	healthHandler := func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	}
	r.GET("/health", healthHandler)
	r.HEAD("health", healthHandler)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
