package services

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/ismailtsdln/netvista/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProber is a mock implementation of ports.Prober
type MockProber struct {
	mock.Mock
}

func (m *MockProber) Probe(ctx context.Context, target domain.Target) (*domain.Metadata, string, error) {
	args := m.Called(ctx, target)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(*domain.Metadata), args.String(1), args.Error(2)
}

// MockRenderer is a mock implementation of ports.Renderer
type MockRenderer struct {
	mock.Mock
}

func (m *MockRenderer) Render(ctx context.Context, target domain.Target) (string, string, string, error) {
	args := m.Called(ctx, target)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockRenderer) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockReporter is a mock implementation of ports.Reporter
type MockReporter struct {
	mock.Mock
}

func (m *MockReporter) Report(ctx context.Context, results []domain.ScanResult) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

func TestScannerService_Scan(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	mockProber := new(MockProber)
	mockRenderer := new(MockRenderer)
	mockReporter := new(MockReporter)

	targets := []domain.Target{
		{URL: "http://example.com"},
	}

	metadata := &domain.Metadata{Title: "Example"}
	mockProber.On("Probe", mock.Anything, targets[0]).Return(metadata, "http://example.com", nil)
	mockRenderer.On("Render", mock.Anything, targets[0]).Return("path/to/img", "hash", "React", nil)
	mockReporter.On("Report", mock.Anything, mock.Anything).Return(nil)

	svc := NewScannerService(mockProber, mockRenderer, nil, mockReporter, domain.Config{Concurrency: 1}, logger)

	err := svc.Scan(ctx, targets)

	assert.NoError(t, err)
	mockProber.AssertExpectations(t)
	mockRenderer.AssertExpectations(t)
	mockReporter.AssertExpectations(t)
}
