package llm

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultHTTPClientConfig(t *testing.T) {
	config := DefaultHTTPClientConfig()

	expected := HTTPClientConfig{
		Timeout:               30 * time.Second,
		MaxIdleConns:          100,
		MaxConnsPerHost:       10,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if config != expected {
		t.Errorf("DefaultHTTPClientConfig() = %+v, expected %+v", config, expected)
	}
}

func TestNewOptimizedHTTPClient(t *testing.T) {
	config := HTTPClientConfig{
		Timeout:               15 * time.Second,
		MaxIdleConns:          50,
		MaxConnsPerHost:       5,
		MaxIdleConnsPerHost:   5,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 500 * time.Millisecond,
	}

	client := NewOptimizedHTTPClient(config)

	if client == nil {
		t.Fatal("NewOptimizedHTTPClient() returned nil")
	}

	if client.Timeout != config.Timeout {
		t.Errorf("Client timeout = %v, expected %v", client.Timeout, config.Timeout)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Client transport is not *http.Transport")
	}

	if transport.MaxIdleConns != config.MaxIdleConns {
		t.Errorf("MaxIdleConns = %d, expected %d", transport.MaxIdleConns, config.MaxIdleConns)
	}

	if transport.MaxConnsPerHost != config.MaxConnsPerHost {
		t.Errorf("MaxConnsPerHost = %d, expected %d", transport.MaxConnsPerHost, config.MaxConnsPerHost)
	}

	if transport.MaxIdleConnsPerHost != config.MaxIdleConnsPerHost {
		t.Errorf("MaxIdleConnsPerHost = %d, expected %d", transport.MaxIdleConnsPerHost, config.MaxIdleConnsPerHost)
	}

	if transport.IdleConnTimeout != config.IdleConnTimeout {
		t.Errorf("IdleConnTimeout = %v, expected %v", transport.IdleConnTimeout, config.IdleConnTimeout)
	}

	if transport.TLSHandshakeTimeout != config.TLSHandshakeTimeout {
		t.Errorf("TLSHandshakeTimeout = %v, expected %v", transport.TLSHandshakeTimeout, config.TLSHandshakeTimeout)
	}

	if transport.ExpectContinueTimeout != config.ExpectContinueTimeout {
		t.Errorf("ExpectContinueTimeout = %v, expected %v", transport.ExpectContinueTimeout, config.ExpectContinueTimeout)
	}

	if !transport.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 should be true")
	}
}

func TestSharedHTTPClient(t *testing.T) {
	if SharedHTTPClient == nil {
		t.Fatal("SharedHTTPClient is nil")
	}

	// Test that multiple calls return the same instance
	client1 := SharedHTTPClient
	client2 := SharedHTTPClient

	if client1 != client2 {
		t.Error("SharedHTTPClient should return the same instance")
	}

	// Verify it has the expected timeout
	expectedTimeout := 30 * time.Second
	if client1.Timeout != expectedTimeout {
		t.Errorf("SharedHTTPClient timeout = %v, expected %v", client1.Timeout, expectedTimeout)
	}

	// Verify transport configuration
	transport, ok := client1.Transport.(*http.Transport)
	if !ok {
		t.Fatal("SharedHTTPClient transport is not *http.Transport")
	}

	if transport.MaxIdleConns != 100 {
		t.Errorf("SharedHTTPClient MaxIdleConns = %d, expected 100", transport.MaxIdleConns)
	}

	if transport.MaxConnsPerHost != 10 {
		t.Errorf("SharedHTTPClient MaxConnsPerHost = %d, expected 10", transport.MaxConnsPerHost)
	}
}

func TestHTTPClientOptimizations(t *testing.T) {
	client := SharedHTTPClient
	transport := client.Transport.(*http.Transport)

	// Test that connection pooling is properly configured
	poolingTests := []struct {
		name     string
		value    int
		expected int
	}{
		{"MaxIdleConns", transport.MaxIdleConns, 100},
		{"MaxConnsPerHost", transport.MaxConnsPerHost, 10},
		{"MaxIdleConnsPerHost", transport.MaxIdleConnsPerHost, 10},
	}

	for _, tt := range poolingTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %d, expected %d", tt.name, tt.value, tt.expected)
			}
		})
	}

	// Test timeout configurations
	timeoutTests := []struct {
		name     string
		value    time.Duration
		expected time.Duration
	}{
		{"Client Timeout", client.Timeout, 30 * time.Second},
		{"IdleConnTimeout", transport.IdleConnTimeout, 90 * time.Second},
		{"TLSHandshakeTimeout", transport.TLSHandshakeTimeout, 10 * time.Second},
		{"ExpectContinueTimeout", transport.ExpectContinueTimeout, 1 * time.Second},
	}

	for _, tt := range timeoutTests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("%s = %v, expected %v", tt.name, tt.value, tt.expected)
			}
		})
	}

	// Test HTTP/2 is enabled
	if !transport.ForceAttemptHTTP2 {
		t.Error("HTTP/2 should be enabled for better performance")
	}
}

