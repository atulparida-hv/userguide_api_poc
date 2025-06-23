package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Config holds application configuration
type Config struct {
	UserGuidePath string
	UserGuideFile string
}

// LoadConfig loads configuration from properties file
func LoadConfig(filename string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Warning: Could not open config file %s, using defaults", filename)
		return config, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "userguide.path":
			config.UserGuidePath = value
		case "userguide.filename":
			config.UserGuideFile = value
		}
	}

	return config, scanner.Err()
}

// FileServiceInterface defines the contract for file download operations
type FileServiceInterface interface {
	DownloadUserGuide() (string, error)
}

// FileService implements FileServiceInterface
type FileService struct {
	basePath      string
	userGuideFile string
	utils         *Utils
}

// NewFileService creates a new file service that implements FileServiceInterface
func NewFileService(basePath, userGuideFile string) FileServiceInterface {
	return &FileService{
		basePath:      basePath,
		userGuideFile: userGuideFile,
		utils:         &Utils{},
	}
}

// DownloadUserGuide validates and returns file path for download using configured filename
func (fs *FileService) DownloadUserGuide() (string, error) {
	// Get filename from configuration instead of parameter
	filename := fs.userGuideFile

	// Validate filename using utils
	cleanFilename, err := fs.utils.ValidateFilename(filename)
	if err != nil {
		return "", err
	}

	// Check file extension
	if !fs.utils.IsAllowedExtension(cleanFilename) {
		ext := strings.ToLower(filepath.Ext(cleanFilename))
		return "", fmt.Errorf("file type not allowed: %s", ext)
	}

	// Check for hidden files
	if strings.HasPrefix(cleanFilename, ".") && filepath.Ext(cleanFilename) == "" {
		return "", fmt.Errorf("hidden files not allowed")
	}

	// Construct full file path
	fullPath := filepath.Join(fs.basePath, cleanFilename)

	// Validate file security
	if !fs.utils.IsFileSecure(fullPath, fs.basePath) {
		return "", fmt.Errorf("file access denied or file not found")
	}

	// Return absolute path
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("unable to resolve file path")
	}

	return absPath, nil
}
