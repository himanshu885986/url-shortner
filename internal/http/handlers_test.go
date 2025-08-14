package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"assignment_infracloud/internal/config"
	"assignment_infracloud/internal/service"
	"assignment_infracloud/internal/storage"
)

func Test_NewServer(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{
		HTTPPort: "8080",
		BaseURL:  "http://localhost:8080",
	}

	server := NewServer(context.Background(), shortener, cfg)
	if server == nil {
		t.Fatal("server returned nil")
	}
	if server.shortener != shortener {
		t.Error("server.shortener was not set correctly")
	}
	if server.cfg != cfg {
		t.Error("server.cfg was not set correctly")
	}
}

func TestServer_HandleShorten_ValidRequest(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{
		HTTPPort: "8080",
		BaseURL:  "http://localhost:8080",
	}
	server := NewServer(context.Background(), shortener, cfg)

	reqBody := shortenRequest{URL: "https://example.com/test"}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleShorten(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleShorten() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp shortenResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.ShortURL == "" {
		t.Error("ShortURL is empty")
	}
	if resp.Code == "" {
		t.Error("Code is empty")
	}
	if resp.ShortURL != cfg.BaseURL+"/"+resp.Code {
		t.Errorf("ShortURL = %s, want %s", resp.ShortURL, cfg.BaseURL+"/"+resp.Code)
	}
}

func TestServer_HandleShorten_InvalidMethod(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	methods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/shorten", nil)
			w := httptest.NewRecorder()

			server.handleShorten(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("handleShorten() with %s status = %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestServer_HandleShorten_InvalidJSON(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleShorten(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleShorten() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestServer_HandleShorten_InvalidURL(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	reqBody := shortenRequest{URL: "not-a-valid-url"}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.handleShorten(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("handleShorten() status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestServer_HandleShorten_DuplicateURL(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	url := "https://example.com/duplicate"
	reqBody := shortenRequest{URL: url}
	reqJSON, _ := json.Marshal(reqBody)

	// First request
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBuffer(reqJSON))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	server.handleShorten(w1, req1)

	var resp1 shortenResponse
	json.NewDecoder(w1.Body).Decode(&resp1)

	// Second request with same URL
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBuffer(reqJSON))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	server.handleShorten(w2, req2)

	var resp2 shortenResponse
	json.NewDecoder(w2.Body).Decode(&resp2)

	// Should return same code
	if resp1.Code != resp2.Code {
		t.Errorf("Duplicate URL returned different codes: %s vs %s", resp1.Code, resp2.Code)
	}
}

func TestServer_HandleMetrics_ValidRequest(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	// Pre-populate with some URLs to create domain statistics
	urls := []string{
		"https://youtube.com/watch?v=123",
		"https://youtube.com/watch?v=456",
		"https://stackoverflow.com/questions/123",
		"https://wikipedia.org/wiki/Go",
		"https://wikipedia.org/wiki/Python",
		"https://udemy.com/course/go",
		"https://udemy.com/course/python",
		"https://udemy.com/course/javascript",
	}

	for i, url := range urls {
		code := fmt.Sprintf("code%d", i)
		store.SaveMapping(code, url)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	w := httptest.NewRecorder()

	server.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleMetrics() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp metricsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.TopDomains) != 3 {
		t.Errorf("Expected 3 top domains, got %d", len(resp.TopDomains))
	}

	// Verify the domains are sorted by count (descending)
	if resp.TopDomains[0].Count < resp.TopDomains[1].Count {
		t.Error("Domains are not sorted by count in descending order")
	}
	if resp.TopDomains[1].Count < resp.TopDomains[2].Count {
		t.Error("Domains are not sorted by count in descending order")
	}

	// Verify specific counts
	expectedCounts := map[string]int{
		"udemy.com":     3,
		"youtube.com":   2,
		"wikipedia.org": 2,
	}

	for _, domain := range resp.TopDomains {
		if expectedCount, exists := expectedCounts[domain.Domain]; exists {
			if domain.Count != expectedCount {
				t.Errorf("Domain %s has count %d, expected %d", domain.Domain, domain.Count, expectedCount)
			}
		}
	}
}

func TestServer_HandleMetrics_EmptyStats(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	w := httptest.NewRecorder()

	server.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleMetrics() status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp metricsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(resp.TopDomains) != 0 {
		t.Errorf("Expected 0 top domains for empty stats, got %d", len(resp.TopDomains))
	}
}

func TestServer_HandleMetrics_InvalidMethod(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/v1/metrics", nil)
			w := httptest.NewRecorder()

			server.handleMetrics(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("handleMetrics() with %s status = %d, want %d", method, w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestServer_HandleResolve_ValidCode(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	ctx := context.Background()
	server := NewServer(ctx, shortener, cfg)

	// First create a short URL
	originalURL := "https://example.com/resolve"
	code, _ := shortener.Shorten(ctx, originalURL)

	// Test resolve
	req := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	w := httptest.NewRecorder()

	server.handleResolve(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("handleResolve() status = %d, want %d", w.Code, http.StatusFound)
	}

	location := w.Header().Get("Location")
	if location != originalURL {
		t.Errorf("handleResolve() Location = %s, want %s", location, originalURL)
	}
}

func TestServer_ServeHTTP(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	// Test shorten endpoint
	reqBody := shortenRequest{URL: "https://example.com/test"}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewBuffer(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ServeHTTP() status = %d, want %d", w.Code, http.StatusOK)
	}

	// Test metrics endpoint
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	w2 := httptest.NewRecorder()

	server.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("ServeHTTP() for metrics status = %d, want %d", w2.Code, http.StatusOK)
	}

	// Test resolve endpoint
	code, _ := shortener.Shorten(context.Background(), "https://example.com/resolve")
	req3 := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	w3 := httptest.NewRecorder()

	server.ServeHTTP(w3, req3)

	if w3.Code != http.StatusFound {
		t.Errorf("ServeHTTP() for resolve status = %d, want %d", w3.Code, http.StatusFound)
	}
}

func BenchmarkServer_HandleShorten(b *testing.B) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	reqBody := shortenRequest{URL: "https://example.com/benchmark"}
	reqJSON, _ := json.Marshal(reqBody)
	buf := bytes.NewBuffer(reqJSON)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", buf)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.handleShorten(w, req)
	}
}

func BenchmarkServer_HandleMetrics(b *testing.B) {
	store := storage.NewInMemoryStore()
	shortener := service.NewInMemoryShortener(store)
	cfg := config.Config{HTTPPort: "8080", BaseURL: "http://localhost:8080"}
	server := NewServer(context.Background(), shortener, cfg)

	// Pre-populate with some domains
	for i := 0; i < 100; i++ {
		code := fmt.Sprintf("code%d", i)
		url := fmt.Sprintf("https://example%d.com", i%10)
		store.SaveMapping(code, url)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
		w := httptest.NewRecorder()
		server.handleMetrics(w, req)
	}
}
