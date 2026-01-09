package ports

import (
	"context"

	"github.com/ismailtsdln/netvista/internal/core/domain"
)

// Prober defines the interface for initial HTTP discovery and metadata extraction.
type Prober interface {
	Probe(ctx context.Context, target domain.Target) (*domain.Metadata, string, error)
}

// Renderer defines the interface for capturing visual snapshots of targets.
type Renderer interface {
	Render(ctx context.Context, target domain.Target) (string, string, string, error)
	Close() error
}

// Analyzer defines the interface for analyzing targets for specific attributes (WAF, Fingerprint, etc.).
type Analyzer interface {
	Analyze(ctx context.Context, result *domain.ScanResult) error
	Name() string
}

// Reporter defines the interface for generating scan reports.
type Reporter interface {
	Report(ctx context.Context, results []domain.ScanResult) error
}

// Storage defines the interface for persisting scan data.
type Storage interface {
	Save(ctx context.Context, result domain.ScanResult) error
	Load(ctx context.Context, id string) (*domain.ScanResult, error)
}
