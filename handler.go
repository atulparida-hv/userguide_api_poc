package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
)

// FileHandler handles HTTP requests
type FileHandler struct {
	fileService FileServiceInterface
	utils       *Utils
}

// NewFileHandler creates a new file handler
func NewFileHandler(fileService FileServiceInterface) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		utils:       &Utils{},
	}
}

// RegisterRoutes registers all handler routes with the router
func (fh *FileHandler) RegisterRoutes(r *mux.Router) {
	// Main user guide download route
	r.HandleFunc("/download/userguide", fh.DownloadUserGuideHandler).Methods("GET")

	// Health check route
	r.HandleFunc("/health", fh.HealthCheckHandler).Methods("GET")
}

// DownloadUserGuideHandler handles the /download/userguide route specifically
func (fh *FileHandler) DownloadUserGuideHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("User guide download request from %s", r.RemoteAddr)

	// Service-level security validation (gets filename from config)
	filePath, err := fh.fileService.DownloadUserGuide()
	if err != nil {
		log.Printf("User guide download failed from %s: %s", r.RemoteAddr, err.Error())
		http.Error(w, "User guide not available", http.StatusNotFound)
		return
	}

	// Set content type using utils
	safeFilename := filepath.Base(filePath)
	w.Header().Set("Content-Type", fh.utils.GetContentType(safeFilename))

	// Set content disposition with proper escaping
	escapedFilename := fh.utils.EscapeForHeader(safeFilename)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+escapedFilename+"\"")

	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	log.Printf("Serving user guide: %s to %s", safeFilename, r.RemoteAddr)

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// HealthCheckHandler handles health check requests
func (fh *FileHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"healthy\"}"))
}
