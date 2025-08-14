package service

import (
	"context"
	"fmt"
	"testing"

	"assignment_infracloud/internal/storage"

	"gotest.tools/assert"
)

func TestNewInMemoryShortener(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)

	if shortener == nil {
		t.Fatal("NewInMemoryShortener() returned nil")
	}
}

func TestInMemoryShortener_Shorten_ValidURL(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)
	ctx := context.Background()

	url := "https://example.com/test"
	code, err := shortener.Shorten(ctx, url)

	if err != nil {
		t.Errorf("Shorten() error = %v", err)
	}
	if code == "" {
		t.Error("Shorten() returned empty code")
	}
	assert.Equal(t, code, "1")
}

func TestInMemoryShortener_Shorten_InvalidURL(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)
	ctx := context.Background()

	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com",
		"http://",
		"https://",
		"",
		"https://",
	}

	for _, invalidURL := range invalidURLs {
		t.Run(invalidURL, func(t *testing.T) {
			_, err := shortener.Shorten(ctx, invalidURL)
			if err != ErrInvalidURL {
				t.Errorf("Shorten(%s) error = %v, want %v", invalidURL, err, ErrInvalidURL)
			}
		})
	}
}

func TestInMemoryShortener_Shorten_DuplicateURL(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)
	ctx := context.Background()

	url := "https://example.com/duplicate"

	// First shortening
	code1, err := shortener.Shorten(ctx, url)
	if err != nil {
		t.Fatalf("First Shorten() error = %v", err)
	}

	// Second shortening of same URL
	code2, err := shortener.Shorten(ctx, url)
	if err != nil {
		t.Fatalf("Second Shorten() error = %v", err)
	}

	// Should return same code
	if code1 != code2 {
		t.Errorf("Duplicate URL shortening returned different codes: %s vs %s", code1, code2)
	}
}

func TestInMemoryShortener_Resolve_ValidCode(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)
	ctx := context.Background()

	originalURL := "https://example.com/resolve"
	code, err := shortener.Shorten(ctx, originalURL)
	if err != nil {
		t.Fatalf("Shorten() error = %v", err)
	}

	resolvedURL, err := shortener.Resolve(ctx, code)
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}
	if resolvedURL != originalURL {
		t.Errorf("Resolve() = %v, want %v", resolvedURL, originalURL)
	}
}

func TestInMemoryShortener_Resolve_InvalidCode(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)
	ctx := context.Background()

	_, err := shortener.Resolve(ctx, "nonexistent")
	if err != storage.ErrNotFound {
		t.Errorf("Resolve() error = %v, want %v", err, storage.ErrNotFound)
	}
}

func TestInMemoryShortener_ShortenAndResolve_Flow(t *testing.T) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)
	ctx := context.Background()

	urls := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://google.com",
		"https://github.com",
	}

	codes := make([]string, len(urls))

	// Shorten all URLs
	for i, url := range urls {
		code, err := shortener.Shorten(ctx, url)
		if err != nil {
			t.Fatalf("Shorten(%s) error = %v", url, err)
		}
		codes[i] = code
	}

	// Verify all codes are unique
	codeMap := make(map[string]bool)
	for _, code := range codes {
		if codeMap[code] {
			t.Errorf("Duplicate code generated: %s", code)
		}
		codeMap[code] = true
	}

	// Resolve all codes
	for i, code := range codes {
		resolvedURL, err := shortener.Resolve(ctx, code)
		if err != nil {
			t.Errorf("Resolve(%s) error = %v", code, err)
		}
		if resolvedURL != urls[i] {
			t.Errorf("Resolve(%s) = %v, want %v", code, resolvedURL, urls[i])
		}
	}
}

func TestIsValidURL(t *testing.T) {
	validURLs := []string{
		"https://example.com",
		"http://example.com",
		"https://example.com/path",
		"https://example.com/path?param=value",
		"https://example.com/path#fragment",
		"http://localhost:8080",
		"https://subdomain.example.com",
	}

	invalidURLs := []string{
		"not-a-url",
		"ftp://example.com",
		"file:///path/to/file",
		"http://",
		"https://",
		"",
		"https://",
		"://example.com",
	}

	for _, url := range validURLs {
		t.Run("valid_"+url, func(t *testing.T) {
			if !isValidURL(url) {
				t.Errorf("isValidURL(%s) = false, want true", url)
			}
		})
	}

	for _, url := range invalidURLs {
		t.Run("invalid_"+url, func(t *testing.T) {
			if isValidURL(url) {
				t.Errorf("isValidURL(%s) = true, want false", url)
			}
		})
	}
}

func BenchmarkInMemoryShortener_Shorten(b *testing.B) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := fmt.Sprintf("https://example%d.com", i)
		shortener.Shorten(ctx, url)
	}
}

func BenchmarkInMemoryShortener_Resolve(b *testing.B) {
	store := storage.NewInMemoryStore()
	shortener := NewInMemoryShortener(store)
	ctx := context.Background()

	// Pre-populate with some URLs
	urls := make([]string, 100)
	codes := make([]string, 100)
	for i := 0; i < 100; i++ {
		url := fmt.Sprintf("https://example%d.com", i)
		code, _ := shortener.Shorten(ctx, url)
		urls[i] = url
		codes[i] = code
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := codes[i%len(codes)]
		shortener.Resolve(ctx, code)
	}
}
