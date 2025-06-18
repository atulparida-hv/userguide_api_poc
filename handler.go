package main

import (
	"net/http"
	"path/filepath"
)

func sendPDF(w http.ResponseWriter, filePath string) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"user-guide.pdf\"")
	http.ServeFile(w, nil, filepath.Clean(filePath))
}

func PublicDownloadHandler(w http.ResponseWriter, r *http.Request) {
	sendPDF(w, "./static/user-guide.pdf")
}

func ProtectedDownloadHandler(w http.ResponseWriter, r *http.Request) {
	sendPDF(w, "./static/user-guide.pdf")
}
