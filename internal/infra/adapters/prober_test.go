package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ismailtsdln/netvista/internal/core/domain"
	"github.com/stretchr/testify/assert"
)

func TestProberAdapter_Probe(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("<html><title>Test Page</title><body>Hello World</body></html>"))
	}))
	defer server.Close()

	adapter := NewProberAdapter(2*time.Second, "", nil)
	target := domain.Target{URL: server.URL}

	metadata, resolvedURL, err := adapter.Probe(context.Background(), target)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, "Test Page", metadata.Title)
	assert.Equal(t, http.StatusOK, metadata.StatusCode)
	assert.Contains(t, resolvedURL, "http://")
}
