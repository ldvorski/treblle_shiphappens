package main

import (
	"log"
	"treblle_project/internal/database"
	"treblle_project/internal/handlers"
	"treblle_project/internal/jikan"
	"treblle_project/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	db, err := database.New("./api_monitor.db")
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
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
