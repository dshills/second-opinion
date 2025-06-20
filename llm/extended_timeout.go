package llm

import (
	"net/http"
	"time"
)

// ExtendedHTTPClientConfig returns configuration optimized for large code reviews
func ExtendedHTTPClientConfig() HTTPClientConfig {
	return HTTPClientConfig{
		Timeout:               5 * time.Minute, // Increased from 30s to 5 minutes
		MaxIdleConns:          100,
		MaxConnsPerHost:       10,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// LargeReviewHTTPClient provides an HTTP client optimized for large code reviews
var LargeReviewHTTPClient = NewOptimizedHTTPClient(ExtendedHTTPClientConfig())

// ProviderWithExtendedTimeout wraps a provider to use extended timeout for large reviews
type ProviderWithExtendedTimeout struct {
	Provider
	httpClient *http.Client
}

// NewProviderWithExtendedTimeout creates a provider wrapper with extended timeout
func NewProviderWithExtendedTimeout(provider Provider) *ProviderWithExtendedTimeout {
	return &ProviderWithExtendedTimeout{
		Provider:   provider,
		httpClient: LargeReviewHTTPClient,
	}
}
