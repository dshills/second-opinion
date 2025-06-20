package llm

import (
	"context"
	"fmt"
)

// MockProvider is a simple mock provider for testing
type MockProvider struct {
	ProviderName string
	Response     string
	Error        error
	CalledWith   string
	CalledCount  int
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		ProviderName: name,
		Response:     "Mock analysis response",
	}
}

// Analyze implements the Provider interface
func (m *MockProvider) Analyze(ctx context.Context, prompt string) (string, error) {
	m.CalledWith = prompt
	m.CalledCount++

	if m.Error != nil {
		return "", m.Error
	}

	// Return a simple response based on the prompt
	if m.Response != "" {
		return m.Response, nil
	}

	return fmt.Sprintf("Mock %s analysis of: %s", m.ProviderName, prompt[:min(50, len(prompt))]), nil
}

// Name implements the Provider interface
func (m *MockProvider) Name() string {
	return m.ProviderName
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
