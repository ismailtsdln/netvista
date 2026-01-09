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
	Render(ctx context.Context, target domain.Target) (string, string, error) // Returns screenshot path and pHash
	Close() error
}

// Analyzer defines the interface for post-scan analysis (Fingerprinting, WAF detection).
type Analyzer interface {
	Analyze(ctx context.Context, result *domain.ScanResult) error
	Name() string
}

// Reporter defines the interface for generating output files (HTML, JSON, CSV).
type Reporter interface {
	Write(results []domain.ScanResult) error
}

// Storage defines the interface for persistent data (Signatures, Config).
type Storage interface {
	LoadSignatures() ([]interface{}, error) // Generic for now
	SaveResult(result domain.ScanResult) error
}
