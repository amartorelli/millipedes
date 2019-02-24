package fetcher

import (
	"bytes"
	"fmt"
	"io"
)

// MockFetcher mocks a fetcher
type MockFetcher struct {
	websites map[string][]byte
}

// NewMockFetcher returns a new MockFetcher
func NewMockFetcher(websites map[string][]byte) *MockFetcher {
	return &MockFetcher{websites: websites}
}

// Fetch returns fake data
func (f *MockFetcher) Fetch(url string) (io.Reader, error) {
	body, ok := f.websites[url]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return bytes.NewReader(body), nil
}
