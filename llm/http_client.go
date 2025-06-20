package llm

import (
	"net/http"
	"time"
)

// HTTPClientConfig holds configuration for HTTP client optimization
type HTTPClientConfig struct {
	Timeout               time.Duration
	MaxIdleConns          int
	MaxConnsPerHost       int
	MaxIdleConnsPerHost   int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
}

// DefaultHTTPClientConfig returns optimized defaults for LLM API calls
func DefaultHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:               5 * time.Minute, // Increased from 30s to 5 minutes for large reviews
		MaxIdleConns:          100,
		MaxConnsPerHost:       10,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// NewOptimizedHTTPClient creates an HTTP client optimized for API calls
func NewOptimizedHTTPClient(config HTTPClientConfig) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:          config.MaxIdleConns,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		// Enable HTTP/2
		ForceAttemptHTTP2: true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}
}

// SharedHTTPClient provides a singleton HTTP client optimized for LLM API calls
var SharedHTTPClient = NewOptimizedHTTPClient(DefaultHTTPClientConfig())
