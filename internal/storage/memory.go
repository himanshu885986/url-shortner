package storage

import (
	"errors"
	"net/url"
	"sort"
	"sync"
)

var ErrNotFound = errors.New("not found")

type DomainStats struct {
	Domain string
	Count  int
}

type InMemoryStore struct {
	mu           sync.RWMutex
	idCounter    uint64
	codeToURL    map[string]string
	urlToCode    map[string]string
	domainCounts map[string]int
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		codeToURL:    make(map[string]string),
		urlToCode:    make(map[string]string),
		domainCounts: make(map[string]int),
	}
}

func (s *InMemoryStore) NextID() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.idCounter++
	return s.idCounter
}

func (s *InMemoryStore) SaveMapping(code, url string) {
	s.mu.Lock()
	s.codeToURL[code] = url
	s.urlToCode[url] = code

	// Track domain statistics
	if domain := extractDomain(url); domain != "" {
		s.domainCounts[domain]++
	}
	s.mu.Unlock()
}

func (s *InMemoryStore) GetURL(code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.codeToURL[code]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *InMemoryStore) GetCode(url string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	code, ok := s.urlToCode[url]
	if !ok {
		return "", ErrNotFound
	}
	return code, nil
}

func (s *InMemoryStore) GetTopDomains(limit int) []DomainStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice for sorting
	stats := make([]DomainStats, 0, len(s.domainCounts))
	for domain, count := range s.domainCounts {
		stats = append(stats, DomainStats{Domain: domain, Count: count})
	}

	// Sort by count in descending order
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			// If counts are equal, sort alphabetically by domain
			return stats[i].Domain < stats[j].Domain
		}
		return stats[i].Count > stats[j].Count
	})

	// Return top N domains
	if limit > len(stats) {
		limit = len(stats)
	}
	return stats[:limit]
}

func extractDomain(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}
