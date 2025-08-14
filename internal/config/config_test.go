package config

import (
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("PORT")
	os.Unsetenv("BASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPPort != "8080" {
		t.Errorf("Load().HTTPPort = %v, want %v", cfg.HTTPPort, "8080")
	}

	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("Load().BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:8080")
	}
}

func TestLoad_CustomPort(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Unsetenv("BASE_URL")
	defer os.Unsetenv("PORT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPPort != "9090" {
		t.Errorf("Load().HTTPPort = %v, want %v", cfg.HTTPPort, "9090")
	}

	if cfg.BaseURL != "http://localhost:9090" {
		t.Errorf("Load().BaseURL = %v, want %v", cfg.BaseURL, "http://localhost:9090")
	}
}

func TestLoad_CustomBaseURL(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("BASE_URL", "https://myshortener.com")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("BASE_URL")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPPort != "9090" {
		t.Errorf("Load().HTTPPort = %v, want %v", cfg.HTTPPort, "9090")
	}

	if cfg.BaseURL != "https://myshortener.com" {
		t.Errorf("Load().BaseURL = %v, want %v", cfg.BaseURL, "https://myshortener.com")
	}
}

func TestLoad_InvalidBaseURL(t *testing.T) {
	os.Setenv("BASE_URL", "not-a-valid-url")
	defer os.Unsetenv("BASE_URL")

	_, err := Load()
	if err == nil {
		t.Error("Load() should return error for invalid BASE_URL")
	}
}
