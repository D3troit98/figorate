package main

import (
	"log"
	"os"

	"figorate/database"
	"figorate/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to MongoDB
	database.ConnectDatabase()
	defer database.DisconnectDatabase()

	// Set up Gin router
	router := gin.Default()

	// Initialize routes
	routes.SetupAuthRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
