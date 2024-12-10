package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"figorate/database"
	"figorate/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Add this function to create the home page template
func createHomePageTemplate() *template.Template {
	tmpl, err := template.New("home").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Figorate API Service</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            line-height: 1.6;
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #ddd;
            padding-bottom: 10px;
        }
        .status {
            background-color: #e7f3fe;
            border-left: 6px solid #2196F3;
            padding: 10px;
            margin: 20px 0;
        }
        .routes {
            background-color: #f1f1f1;
            padding: 15px;
            border-radius: 5px;
        }
        .route-group {
            margin-bottom: 15px;
        }
        .route-group h3 {
            margin-bottom: 10px;
            color: #444;
        }
        .route-item {
            background-color: #fff;
            border: 1px solid #ddd;
            padding: 10px;
            margin-bottom: 5px;
            border-radius: 3px;
        }
    </style>
</head>
<body>
    <h1>Figorate API Service</h1>

    <div class="status">
        <h2>Service Status</h2>
        <p>✅ Service is running and healthy</p>
        <p>Current Time: {{ .CurrentTime }}</p>
        <p>Uptime: {{ .Uptime }}</p>
    </div>

    <div class="routes">
        <h2>Available API Routes</h2>

        <div class="route-group">
            <h3>Authentication Routes</h3>
            <div class="route-item">POST /signup - User Registration</div>
            <div class="route-item">POST /signin - User Login</div>
            <div class="route-item">POST /refresh-token - Refresh Authentication Token</div>
            <div class="route-item">GET /verify-email - Verify User Email</div>
            <div class="route-item">GET /profile - Get User Profile (Protected)</div>
            <div class="route-item">POST /onboarding - Complete User Onboarding (Protected)</div>
        </div>

        <div class="route-group">
            <h3>Quote Routes</h3>
            <div class="route-item">GET /qoutes - Get Quotes</div>
            <div class="route-item">GET /qoutes/random - Get Random Quote</div>
            <div class="route-item">GET /qoutes/:id - Get Specific Quote</div>
            <div class="route-item">POST /qoutes - Create Quote (Protected)</div>
        </div>

        <div class="route-group">
            <h3>Meal Routes</h3>
            <div class="route-item">POST /meals/add - Add Meal (Protected)</div>
            <div class="route-item">POST /meals/generate-plan - Generate Meal Plan (Protected)</div>
            <div class="route-item">GET /meals/plan/:day - Get Daily Meal Plan (Protected)</div>
            <div class="route-item">POST /meals/recalibrate - Recalibrate Meal Plan (Protected)</div>
        </div>
    </div>
</body>
</html>
    `)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}
	return tmpl
}

// Add a new function to handle the home route
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := createHomePageTemplate()

	data := struct {
		CurrentTime string
		Uptime      string
	}{
		CurrentTime: time.Now().Format(time.RFC1123),
		Uptime:      time.Since(startTime).Round(time.Second).String(),
	}

	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var (
	startTime time.Time
)

func main() {
	// Log current working directory
	startTime = time.Now()
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
	router.GET("/", func(c *gin.Context) {
		homeHandler(c.Writer, c.Request)
	})
	log.Fatal(router.Run(":" + port))
}
