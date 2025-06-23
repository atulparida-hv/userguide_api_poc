package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	config, err := LoadConfig("application.properties")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	if config.UserGuidePath == "" {
		log.Fatal("User guide path cannot be empty")
	}

	// Create directory if needed
	if _, err := os.Stat(config.UserGuidePath); os.IsNotExist(err) {
		err := os.MkdirAll(config.UserGuidePath, 0755)
		if err != nil {
			log.Fatal("Failed to create userguides directory:", err)
		}
	}

	// Initialize service with interface
	var fileService FileServiceInterface = NewFileService(config.UserGuidePath, config.UserGuideFile)
	fileHandler := NewFileHandler(fileService)
	// Create router
	r := mux.NewRouter()
	r.Use(securityMiddleware)

	// Register routes using handler method
	fileHandler.RegisterRoutes(r)

	log.Printf("Server starting on port %s", "8080")
	log.Printf("User guides directory: %s", config.UserGuidePath)
	log.Printf("Configured user guide file: %s", config.UserGuideFile)
	log.Println("Available endpoints:")
	log.Println("  GET /download/userguide - Download configured user guide")
	log.Println("  GET /health - Health check")

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
