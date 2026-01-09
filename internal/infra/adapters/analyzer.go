package adapters

import (
	"context"

	"github.com/ismailtsdln/netvista/internal/core/domain"
	"github.com/ismailtsdln/netvista/internal/plugins"
	"github.com/ismailtsdln/netvista/pkg/models"
)

// WafAnalyzerAdapter wraps the WafPlugin.
type WafAnalyzerAdapter struct {
	p *plugins.WafPlugin
}

// NewWafAnalyzerAdapter creates a new Waf analyzer adapter.
func NewWafAnalyzerAdapter(p *plugins.WafPlugin) *WafAnalyzerAdapter {
	return &WafAnalyzerAdapter{p: p}
}

// Analyze performs WAF detection on the result.
func (a *WafAnalyzerAdapter) Analyze(ctx context.Context, result *domain.ScanResult) error {
	// Map to legacy model for plugin compatibility
	legacyTarget := &models.Target{
		URL: result.Target.URL,
		Metadata: models.ResponseMetadata{
			Headers:    result.Metadata.Headers,
			Technology: result.Metadata.Technology,
		},
	}

	if err := a.p.Execute(legacyTarget); err != nil {
		return err
	}

	// Map back
	result.Metadata.Technology = legacyTarget.Metadata.Technology
	return nil
}

// Name returns the analyzer name.
func (a *WafAnalyzerAdapter) Name() string {
	return a.p.Name()
}
