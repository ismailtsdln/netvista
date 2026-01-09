package adapters

import (
	"context"
	"time"

	"github.com/ismailtsdln/netvista/internal/core/domain"
	"github.com/ismailtsdln/netvista/internal/prober"
)

// ProberAdapter wraps the existing prober logic.
type ProberAdapter struct {
	p *prober.Prober
}

// NewProberAdapter creates a new prober adapter.
func NewProberAdapter(timeout time.Duration, proxy string, headers map[string]string) *ProberAdapter {
	return &ProberAdapter{
		p: prober.NewProber(timeout, proxy, headers, 10), // Default 10 redirects
	}
}

// Probe extracts metadata from a target.
func (a *ProberAdapter) Probe(ctx context.Context, target domain.Target) (*domain.Metadata, string, error) {
	res, err := a.p.Probe(ctx, target.URL)
	if err != nil {
		return nil, "", err
	}

	return &domain.Metadata{
		Title:      res.Metadata.Title,
		StatusCode: res.Metadata.StatusCode,
		Technology: res.Metadata.Technology,
		Headers:    res.Metadata.Headers,
		ContentLen: res.Metadata.ContentLen,
		Timestamp:  res.Metadata.Timestamp,
		Redirects:  res.Metadata.Redirects,
	}, res.URL, nil
}
