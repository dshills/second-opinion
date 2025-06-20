package llm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"time"
)

// RetryConfig holds configuration for retry logic
type RetryConfig struct {
	MaxRetries      int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	BackoffMultiple float64
}

// DefaultRetryConfig returns sensible defaults for retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		BaseDelay:       1 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffMultiple: 2.0,
	}
}

// IsRetryableError determines if an error should trigger a retry
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Context deadline exceeded or canceled should not be retried
	if err == context.DeadlineExceeded || err == context.Canceled {
		return false
	}

	// Network errors are retryable
	if _, ok := err.(net.Error); ok {
		return true
	}

	return false
}

// IsRetryableHTTPStatus determines if an HTTP status code should trigger a retry
func IsRetryableHTTPStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests: // 429 - Rate limit
		return true
	case http.StatusInternalServerError: // 500
		return true
	case http.StatusBadGateway: // 502
		return true
	case http.StatusServiceUnavailable: // 503
		return true
	case http.StatusGatewayTimeout: // 504
		return true
	default:
		return false
	}
}

// CalculateDelay calculates the delay for a retry attempt using exponential backoff
func (rc RetryConfig) CalculateDelay(attempt int) time.Duration {
	if attempt == 0 {
		return rc.BaseDelay
	}

	delay := float64(rc.BaseDelay) * math.Pow(rc.BackoffMultiple, float64(attempt))

	// Add jitter (±25% random variation)
	jitter := 0.25 * delay * (2*float64(time.Now().UnixNano()%1000)/1000 - 1)
	delay += jitter

	delayDuration := time.Duration(delay)
	if delayDuration > rc.MaxDelay {
		delayDuration = rc.MaxDelay
	}

	return delayDuration
}

// RetryableHTTPRequest performs an HTTP request with retry logic
func RetryableHTTPRequest(ctx context.Context, client *http.Client, req *http.Request, config RetryConfig) (*http.Response, error) {
	var lastErr error

	// Read the request body once if it exists
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
	}

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Clone the request for retry attempts
		reqCopy := req.Clone(ctx)

		// Set a new body reader for each attempt
		if bodyBytes != nil {
			reqCopy.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err := client.Do(reqCopy)

		// If successful, return immediately
		if err == nil && !IsRetryableHTTPStatus(resp.StatusCode) {
			return resp, nil
		}

		// Check if this error/status is retryable
		shouldRetry := false
		if err != nil {
			shouldRetry = IsRetryableError(err)
			lastErr = err
		} else if resp != nil {
			shouldRetry = IsRetryableHTTPStatus(resp.StatusCode)
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			// Close the response body for failed attempts
			resp.Body.Close()
		}

		// If not retryable, return immediately
		if !shouldRetry {
			if resp != nil && err == nil {
				return resp, nil // Return the response for non-retryable status codes
			}
			return nil, lastErr
		}

		// If this was the last attempt, return the error
		if attempt == config.MaxRetries {
			return nil, fmt.Errorf("request failed after %d attempts: %w", attempt+1, lastErr)
		}

		// Wait before retrying
		delay := config.CalculateDelay(attempt)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}

// RetryableOperation performs a generic operation with retry logic
func RetryableOperation[T any](ctx context.Context, config RetryConfig, operation func() (T, error)) (T, error) {
	var zero T
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		result, err := operation()

		// If successful, return immediately
		if err == nil {
			return result, nil
		}

		// Check if this error is retryable
		if !IsRetryableError(err) {
			return zero, err
		}

		lastErr = err

		// If this was the last attempt, return the error
		if attempt == config.MaxRetries {
			break
		}

		// Wait before retrying
		delay := config.CalculateDelay(attempt)

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}

	return zero, fmt.Errorf("operation failed after %d attempts: %w", config.MaxRetries+1, lastErr)
}
