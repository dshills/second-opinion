package llm

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", config.MaxRetries)
	}
	if config.BaseDelay != 1*time.Second {
		t.Errorf("Expected BaseDelay=1s, got %v", config.BaseDelay)
	}
	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay=30s, got %v", config.MaxDelay)
	}
	if config.BackoffMultiple != 2.0 {
		t.Errorf("Expected BackoffMultiple=2.0, got %f", config.BackoffMultiple)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "network error",
			err:      &net.DNSError{},
			expected: true,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: false,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsRetryableHTTPStatus(t *testing.T) {
	tests := []struct {
		status   int
		expected bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusForbidden, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			result := IsRetryableHTTPStatus(tt.status)
			if result != tt.expected {
				t.Errorf("IsRetryableHTTPStatus(%d) = %v, expected %v", tt.status, result, tt.expected)
			}
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	config := RetryConfig{
		BaseDelay:       1 * time.Second,
		MaxDelay:        10 * time.Second,
		BackoffMultiple: 2.0,
	}

	// Test first attempt
	delay := config.CalculateDelay(0)
	if delay < 750*time.Millisecond || delay > 1250*time.Millisecond {
		t.Errorf("First attempt delay should be around 1s with jitter, got %v", delay)
	}

	// Test max delay cap
	delay = config.CalculateDelay(10) // This should hit the max delay
	if delay > config.MaxDelay {
		t.Errorf("Delay should not exceed MaxDelay, got %v", delay)
	}
}

func TestRetryableHTTPRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	config := RetryConfig{
		MaxRetries:      2,
		BaseDelay:       10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	resp, err := RetryableHTTPRequest(context.Background(), client, req, config)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRetryableHTTPRequest_RetryOnFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("server error"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	config := RetryConfig{
		MaxRetries:      3,
		BaseDelay:       1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	resp, err := RetryableHTTPRequest(context.Background(), client, req, config)
	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableHTTPRequest_ExceedsMaxRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	config := RetryConfig{
		MaxRetries:      2,
		BaseDelay:       1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	resp, err := RetryableHTTPRequest(context.Background(), client, req, config)
	if err == nil {
		if resp != nil {
			defer resp.Body.Close()
		}
		t.Fatal("Expected error after exceeding max retries")
	}

	if !strings.Contains(err.Error(), "failed after 3 attempts") {
		t.Errorf("Error should mention number of attempts, got: %v", err)
	}
}

func TestRetryableHTTPRequest_NonRetryableStatus(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest) // Non-retryable status
		w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	config := RetryConfig{
		MaxRetries:      3,
		BaseDelay:       1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	resp, err := RetryableHTTPRequest(context.Background(), client, req, config)
	if err != nil {
		t.Fatalf("Expected response even with bad status, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable status, got %d", attempts)
	}
}

func TestRetryableHTTPRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate slow response
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	config := RetryConfig{
		MaxRetries:      3,
		BaseDelay:       50 * time.Millisecond,
		MaxDelay:        200 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	resp, err := RetryableHTTPRequest(ctx, client, req, config)
	if err == nil {
		if resp != nil {
			defer resp.Body.Close()
		}
		t.Fatal("Expected error due to context cancellation")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}
}

func TestRetryableOperation_Success(t *testing.T) {
	attempts := 0
	operation := func() (string, error) {
		attempts++
		if attempts < 2 {
			return "", &net.DNSError{} // Retryable network error
		}
		return "success", nil
	}

	config := RetryConfig{
		MaxRetries:      3,
		BaseDelay:       1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	result, err := RetryableOperation(context.Background(), config, operation)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected 'success', got %s", result)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryableOperation_NonRetryableError(t *testing.T) {
	attempts := 0
	operation := func() (string, error) {
		attempts++
		return "", errors.New("non-retryable error")
	}

	config := RetryConfig{
		MaxRetries:      3,
		BaseDelay:       1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffMultiple: 2.0,
	}

	_, err := RetryableOperation(context.Background(), config, operation)
	if err == nil {
		t.Fatal("Expected error")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
	}
}
