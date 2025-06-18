package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicDownload(t *testing.T) {
	req := httptest.NewRequest("GET", "/public/download", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(PublicDownloadHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status 200, got %v", status)
	}
}

func TestProtectedDownload_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/protected/download", nil)
	rr := httptest.NewRecorder()

	handler := AuthMiddleware(http.HandlerFunc(ProtectedDownloadHandler))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestProtectedDownload_Authorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/protected/download", nil)
	req.Header.Set("Authorization", "Bearer valid-oauth-token")
	rr := httptest.NewRecorder()

	handler := AuthMiddleware(http.HandlerFunc(ProtectedDownloadHandler))
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status 200, got %v", status)
	}
}
