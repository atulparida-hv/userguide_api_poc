package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	// Public route example
	r.HandleFunc("/public/download", PublicDownloadHandler).Methods("GET")

	// Protected route
	r.Handle("/protected/download", AuthMiddleware(http.HandlerFunc(ProtectedDownloadHandler))).Methods("GET")

	// Fallback for serving static PDF file
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/static/"))))

	// logging
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
