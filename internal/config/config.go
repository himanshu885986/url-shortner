package config

import (
    "fmt"
    "net/url"
    "os"
)

type Config struct {
    HTTPPort string
    BaseURL  string
}

func Load() (Config, error) {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    baseURL := os.Getenv("BASE_URL")
    if baseURL == "" {
        baseURL = fmt.Sprintf("http://localhost:%s", port)
    }
    if _, err := url.ParseRequestURI(baseURL); err != nil {
        return Config{}, fmt.Errorf("invalid BASE_URL: %w", err)
    }

    return Config{
        HTTPPort: port,
        BaseURL:  baseURL,
    }, nil
}


