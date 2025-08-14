package service

import (
	"context"
	"errors"
	"net/url"

	"assignment_infracloud/internal/encoding"
	"assignment_infracloud/internal/storage"
)

var ErrInvalidURL = errors.New("invalid url")

type Shortener interface {
	Shorten(ctx context.Context, longURL string) (string, error)
	Resolve(ctx context.Context, code string) (string, error)
	GetTopDomains(ctx context.Context, limit int) []storage.DomainStats
}

type InMemoryShortener struct {
	store *storage.InMemoryStore
}

func NewInMemoryShortener(store *storage.InMemoryStore) Shortener {
	return &InMemoryShortener{store: store}
}

func (s *InMemoryShortener) Shorten(ctx context.Context, longURL string) (string, error) {
	if !isValidURL(longURL) {
		return "", ErrInvalidURL
	}
	if code, err := s.store.GetCode(longURL); err == nil {
		return code, nil
	}
	id := s.store.NextID()
	code := encoding.Base62Encode(id)
	s.store.SaveMapping(code, longURL)
	return code, nil
}

func (s *InMemoryShortener) Resolve(ctx context.Context, code string) (string, error) {
	return s.store.GetURL(code)
}

func (s *InMemoryShortener) GetTopDomains(ctx context.Context, limit int) []storage.DomainStats {
	return s.store.GetTopDomains(limit)
}

func isValidURL(u string) bool {
	parsedUrl, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
		return false
	}
	return parsedUrl.Host != ""
}
