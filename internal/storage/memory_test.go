package storage

import (
	"fmt"
	"sync"
	"testing"
)

func TestNewInMemoryStore(t *testing.T) {
	store := NewInMemoryStore()
	if store == nil {
		t.Fatal("NewInMemoryStore() returned nil")
	}
	if store.codeToURL == nil {
		t.Error("codeToURL map was not initialized")
	}
	if store.urlToCode == nil {
		t.Error("urlToCode map was not initialized")
	}
	if store.domainCounts == nil {
		t.Error("domainCounts map was not initialized")
	}
}

func TestInMemoryStore_NextID(t *testing.T) {
	store := NewInMemoryStore()

	// Test that IDs are sequential
	ids := make([]uint64, 10)
	for i := 0; i < 10; i++ {
		ids[i] = store.NextID()
	}

	// Verify IDs are sequential starting from 1
	for i, id := range ids {
		expected := uint64(i + 1)
		if id != expected {
			t.Errorf("NextID() returned %d, want %d", id, expected)
		}
	}
}

func TestInMemoryStore_SaveAndGetMapping(t *testing.T) {
	store := NewInMemoryStore()
	code := "abc123"
	url := "https://example.com"

	// Save mapping
	store.SaveMapping(code, url)

	// Test GetURL
	retrievedURL, err := store.GetURL(code)
	if err != nil {
		t.Errorf("GetURL() error = %v", err)
	}
	if retrievedURL != url {
		t.Errorf("GetURL() = %v, want %v", retrievedURL, url)
	}

	// Test GetCode
	retrievedCode, err := store.GetCode(url)
	if err != nil {
		t.Errorf("GetCode() error = %v", err)
	}
	if retrievedCode != code {
		t.Errorf("GetCode() = %v, want %v", retrievedCode, code)
	}
}

func TestInMemoryStore_GetURL_NotFound(t *testing.T) {
	store := NewInMemoryStore()

	_, err := store.GetURL("nonexistent")
	if err != ErrNotFound {
		t.Errorf("GetURL() error = %v, want %v", err, ErrNotFound)
	}
}

func TestInMemoryStore_GetCode_NotFound(t *testing.T) {
	store := NewInMemoryStore()

	_, err := store.GetCode("https://nonexistent.com")
	if err != ErrNotFound {
		t.Errorf("GetCode() error = %v, want %v", err, ErrNotFound)
	}
}

func TestInMemoryStore_DomainTracking(t *testing.T) {
	store := NewInMemoryStore()

	// Shorten URLs from different domains
	urls := []string{
		"https://youtube.com/watch?v=123",
		"https://youtube.com/watch?v=456",
		"https://youtube.com/watch?v=789",
		"https://stackoverflow.com/questions/123",
		"https://wikipedia.org/wiki/Go",
		"https://wikipedia.org/wiki/Python",
		"https://udemy.com/course/go",
		"https://udemy.com/course/python",
		"https://udemy.com/course/javascript",
		"https://udemy.com/course/react",
		"https://udemy.com/course/nodejs",
		"https://udemy.com/course/docker",
	}

	for i, url := range urls {
		code := fmt.Sprintf("code%d", i)
		store.SaveMapping(code, url)
	}

	// Get top domains
	topDomains := store.GetTopDomains(3)

	// Verify results
	if len(topDomains) != 3 {
		t.Errorf("GetTopDomains(3) returned %d domains, want 3", len(topDomains))
	}

	// Check that Udemy has the highest count (6)
	if topDomains[0].Domain != "udemy.com" || topDomains[0].Count != 6 {
		t.Errorf("Expected top domain to be udemy.com with count 6, got %s with count %d",
			topDomains[0].Domain, topDomains[0].Count)
	}

	// Check that YouTube has the second highest count (3)
	if topDomains[1].Domain != "youtube.com" || topDomains[1].Count != 3 {
		t.Errorf("Expected second domain to be youtube.com with count 3, got %s with count %d",
			topDomains[1].Domain, topDomains[1].Count)
	}

	// Check that Wikipedia has the third highest count (2)
	if topDomains[2].Domain != "wikipedia.org" || topDomains[2].Count != 2 {
		t.Errorf("Expected third domain to be wikipedia.org with count 2, got %s with count %d",
			topDomains[2].Domain, topDomains[2].Count)
	}
}

func TestInMemoryStore_GetTopDomains_Empty(t *testing.T) {
	store := NewInMemoryStore()

	topDomains := store.GetTopDomains(3)
	if len(topDomains) != 0 {
		t.Errorf("GetTopDomains(3) returned %d domains, want 0", len(topDomains))
	}
}

func TestInMemoryStore_GetTopDomains_LimitExceedsAvailable(t *testing.T) {
	store := NewInMemoryStore()

	// Add only 2 domains
	store.SaveMapping("code1", "https://example1.com")
	store.SaveMapping("code2", "https://example2.com")

	topDomains := store.GetTopDomains(5)
	if len(topDomains) != 2 {
		t.Errorf("GetTopDomains(5) returned %d domains, want 2", len(topDomains))
	}
}

func TestInMemoryStore_GetTopDomains_TieBreaking(t *testing.T) {
	store := NewInMemoryStore()

	// Add domains with equal counts but different alphabetical order
	store.SaveMapping("code1", "https://zebra.com")
	store.SaveMapping("code2", "https://zebra.com")
	store.SaveMapping("code3", "https://apple.com")
	store.SaveMapping("code4", "https://apple.com")

	topDomains := store.GetTopDomains(2)

	// Both should have count 2, but apple.com should come first alphabetically
	if topDomains[0].Domain != "apple.com" || topDomains[0].Count != 2 {
		t.Errorf("Expected first domain to be apple.com with count 2, got %s with count %d",
			topDomains[0].Domain, topDomains[0].Count)
	}

	if topDomains[1].Domain != "zebra.com" || topDomains[1].Count != 2 {
		t.Errorf("Expected second domain to be zebra.com with count 2, got %s with count %d",
			topDomains[1].Domain, topDomains[1].Count)
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{"simple domain", "https://example.com", "example.com"},
		{"domain with path", "https://example.com/path", "example.com"},
		{"domain with query", "https://example.com?param=value", "example.com"},
		{"domain with port", "https://example.com:8080", "example.com"},
		{"subdomain", "https://sub.example.com", "sub.example.com"},
		{"www subdomain", "https://www.example.com", "www.example.com"},
		{"invalid url", "not-a-url", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDomain(tt.url)
			if result != tt.expected {
				t.Errorf("extractDomain(%s) = %s, want %s", tt.url, result, tt.expected)
			}
		})
	}
}

func TestInMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryStore()
	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent NextID calls
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			store.NextID()
		}()
	}
	wg.Wait()

	// Verify we got the expected number of IDs
	expectedID := uint64(numGoroutines)
	actualID := store.NextID()
	if actualID != expectedID+1 {
		t.Errorf("After concurrent access, NextID() = %d, want %d", actualID, expectedID+1)
	}
}

func TestInMemoryStore_ConcurrentReadWrite(t *testing.T) {
	store := NewInMemoryStore()
	var wg sync.WaitGroup
	numGoroutines := 50

	// Start writers
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			code := fmt.Sprintf("code%d", id)
			url := fmt.Sprintf("https://example%d.com", id)
			store.SaveMapping(code, url)
		}(i)
	}

	// Start readers
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			code := fmt.Sprintf("code%d", id)
			url := fmt.Sprintf("https://example%d.com", id)

			// Try to read (might not exist yet due to concurrency)
			store.GetURL(code)
			store.GetCode(url)
		}(i)
	}

	wg.Wait()

	// Verify all mappings were saved correctly
	for i := 0; i < numGoroutines; i++ {
		code := fmt.Sprintf("code%d", i)
		url := fmt.Sprintf("https://example%d.com", i)

		retrievedURL, err := store.GetURL(code)
		if err != nil {
			t.Errorf("GetURL(%s) error = %v", code, err)
		}
		if retrievedURL != url {
			t.Errorf("GetURL(%s) = %v, want %v", code, retrievedURL, url)
		}
	}
}

func BenchmarkInMemoryStore_NextID(b *testing.B) {
	store := NewInMemoryStore()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.NextID()
	}
}

func BenchmarkInMemoryStore_SaveMapping(b *testing.B) {
	store := NewInMemoryStore()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := fmt.Sprintf("code%d", i)
		url := fmt.Sprintf("https://example%d.com", i)
		store.SaveMapping(code, url)
	}
}

func BenchmarkInMemoryStore_GetTopDomains(b *testing.B) {
	store := NewInMemoryStore()

	// Pre-populate with some domains
	for i := 0; i < 1000; i++ {
		code := fmt.Sprintf("code%d", i)
		url := fmt.Sprintf("https://example%d.com", i%10) // 10 different domains
		store.SaveMapping(code, url)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetTopDomains(3)
	}
}
