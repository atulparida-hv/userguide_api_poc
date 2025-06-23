package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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

// Security middleware
func securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}
