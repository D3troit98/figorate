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
	// Log current working directory
	currentDir, _ := os.Getwd()
	log.Printf("Current working directory: %s", currentDir)

	// Try to load .env file with more detailed error handling
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading local .env file: %v", err)
	}

	// Log all environment variables for debugging
	for _, env := range os.Environ() {
		log.Println(env)
	}

	// Connect to MongoDB
	database.ConnectDatabase()
	defer database.DisconnectDatabase()

	// Set up Gin router
	router := gin.Default()

	// Initialize routes
	routes.SetupAuthRoutes(router)
	routes.SetupQouteRoutes(router)
	routes.SetupMealRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
