package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config holds application configuration
type Config struct {
	UserGuidePath string
	Port          string
	UserGuideFile string
}

// LoadConfig loads configuration from properties file
func LoadConfig(filename string) (*Config, error) {
	config := &Config{
		UserGuidePath: "./userguides",   // default value
		Port:          "8080",           // default value
		UserGuideFile: "user-guide.pdf", // default value
	}

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
		case "server.port":
			config.Port = value
		case "userguide.filename":
			config.UserGuideFile = value
		}
	}

	return config, scanner.Err()
}

// Utils contains utility methods for file operations
type Utils struct{}

// ValidateFilename validates filename for security
func (u *Utils) ValidateFilename(filename string) (string, error) {
	// URL decode the filename first
	decodedFilename, err := url.QueryUnescape(filename)
	if err != nil {
		return "", fmt.Errorf("invalid filename encoding")
	}

	// Check for null bytes and control characters
	if strings.Contains(decodedFilename, "\x00") {
		return "", fmt.Errorf("null byte detected in filename")
	}

	for _, char := range decodedFilename {
		if char < 32 && char != 9 && char != 10 && char != 13 {
			return "", fmt.Errorf("control character detected in filename")
		}
	}

	// Strict filename pattern validation
	filenamePattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !filenamePattern.MatchString(decodedFilename) {
		return "", fmt.Errorf("filename contains invalid characters")
	}

	// Check filename length
	if len(decodedFilename) > 255 {
		return "", fmt.Errorf("filename too long")
	}

	// Prevent dangerous patterns
	dangerousPatterns := []string{"..", "~/", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	lowerFilename := strings.ToLower(decodedFilename)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerFilename, pattern) {
			return "", fmt.Errorf("dangerous pattern detected in filename: %s", pattern)
		}
	}

	cleanFilename := filepath.Base(decodedFilename)
	if cleanFilename == "" || cleanFilename == "." || cleanFilename == ".." {
		return "", fmt.Errorf("invalid filename after sanitization")
	}

	return cleanFilename, nil
}

// IsAllowedExtension checks if file extension is allowed
func (u *Utils) IsAllowedExtension(filename string) bool {
	allowedExtensions := []string{".pdf", ".doc", ".docx", ".txt", ".md"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowedExt := range allowedExtensions {
		if ext == allowedExt {
			return true
		}
	}
	return false
}

// IsFileSecure validates file exists and is within allowed directory
func (u *Utils) IsFileSecure(fullPath, basePath string) bool {
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return false
	}

	if fileInfo.IsDir() {
		return false
	}

	absBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return false
	}

	absFilePath, err := filepath.Abs(fullPath)
	if err != nil {
		return false
	}

	return strings.HasPrefix(absFilePath, absBasePath+string(filepath.Separator)) || absFilePath == absBasePath
}

// GetContentType returns appropriate content type for file extension
func (u *Utils) GetContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	default:
		return "application/octet-stream"
	}
}

// EscapeForJSON escapes string for safe JSON usage
func (u *Utils) EscapeForJSON(str string) string {
	escaped := strings.ReplaceAll(str, "\\", "\\\\")
	return strings.ReplaceAll(escaped, "\"", "\\\"")
}

// EscapeForHeader escapes string for safe HTTP header usage
func (u *Utils) EscapeForHeader(str string) string {
	return strings.ReplaceAll(str, "\"", "\\\"")
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
