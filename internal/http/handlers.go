package http

import (
	"context"
	"encoding/json"
	"log"
	stdhttp "net/http"

	"assignment_infracloud/internal/config"
	"assignment_infracloud/internal/service"
)

type Server struct {
	mux       *stdhttp.ServeMux
	shortener service.Shortener
	cfg       config.Config
}

func NewServer(ctx context.Context, shortener service.Shortener, cfg config.Config) *Server {
	s := &Server{
		mux:       stdhttp.NewServeMux(),
		shortener: shortener,
		cfg:       cfg,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/v1/shorten", s.handleShorten)
	s.mux.HandleFunc("/api/v1/metrics", s.handleMetrics)
	s.mux.HandleFunc("/", s.handleResolve)
}

func (s *Server) ServeHTTP(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	s.mux.ServeHTTP(w, r)
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

type metricsResponse struct {
	TopDomains []domainStat `json:"top_domains"`
}

type domainStat struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

func (s *Server) handleShorten(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		stdhttp.Error(w, "method not allowed", stdhttp.StatusMethodNotAllowed)
		return
	}
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		stdhttp.Error(w, "invalid json", stdhttp.StatusBadRequest)
		return
	}
	code, err := s.shortener.Shorten(r.Context(), req.URL)
	if err != nil {
		if err == service.ErrInvalidURL {
			stdhttp.Error(w, "invalid url", stdhttp.StatusBadRequest)
			return
		}
		log.Printf("shorten error: %v", err)
		stdhttp.Error(w, "internal error", stdhttp.StatusInternalServerError)
		return
	}
	resp := shortenResponse{
		ShortURL: s.cfg.BaseURL + "/" + code,
		Code:     code,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleMetrics(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodGet {
		stdhttp.Error(w, "method not allowed", stdhttp.StatusMethodNotAllowed)
		return
	}

	domainStats := s.shortener.GetTopDomains(r.Context(), 3)

	// Convert to response format
	topDomains := make([]domainStat, len(domainStats))
	for i, stat := range domainStats {
		topDomains[i] = domainStat{
			Domain: stat.Domain,
			Count:  stat.Count,
		}
	}

	resp := metricsResponse{
		TopDomains: topDomains,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleResolve(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.URL.Path == "/" {
		w.WriteHeader(stdhttp.StatusOK)
		w.Write([]byte("URL Shortener Service Healthcheck service is running"))
		return
	}
	code := r.URL.Path[1:]
	longURL, err := s.shortener.Resolve(r.Context(), code)
	if err != nil {
		stdhttp.NotFound(w, r)
		return
	}
	stdhttp.Redirect(w, r, longURL, 301)
}
