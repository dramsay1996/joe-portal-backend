package main

import (
	"log"
	"net/http"
	"os"

	"joe-portal/backend/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Set up logging to file
	logFile, err := os.OpenFile("/var/log/joe-portal.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer logFile.Close()

	// Set up logging to both file and stdout
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Log environment variables (excluding sensitive ones)
	log.Printf("Server starting with configuration:")
	log.Printf("PORT: %s", os.Getenv("PORT"))
	log.Printf("CORS_ALLOWED_ORIGINS: %s", os.Getenv("CORS_ALLOWED_ORIGINS"))
	log.Printf("FRONTEND_URL: %s", os.Getenv("FRONTEND_URL"))

	// Setup routes
	router := routes.SetupRoutes()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
