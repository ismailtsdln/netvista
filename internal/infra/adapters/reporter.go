package adapters

import (
	"fmt"
	"path/filepath"

	"github.com/ismailtsdln/netvista/internal/core/domain"
	"github.com/ismailtsdln/netvista/internal/report"
	"github.com/ismailtsdln/netvista/pkg/models"
)

// ReporterAdapter wraps the existing reporting logic.
type ReporterAdapter struct {
	outputPath string
}

// NewReporterAdapter creates a new reporter adapter.
func NewReporterAdapter(outputPath string) *ReporterAdapter {
	return &ReporterAdapter{
		outputPath: outputPath,
	}
}

// Write generates all report formats.
func (a *ReporterAdapter) Write(results []domain.ScanResult) error {
	// Map domain results back to legacy models for report compatibility
	var legacyResults []models.Target
	for _, res := range results {
		legacyResults = append(legacyResults, models.Target{
			URL: res.Target.URL,
			Metadata: models.ResponseMetadata{
				Title:      res.Metadata.Title,
				StatusCode: res.Metadata.StatusCode,
				Technology: res.Metadata.Technology,
				Headers:    res.Metadata.Headers,
				ContentLen: res.Metadata.ContentLen,
				Redirects:  res.Metadata.Redirects,
				Timestamp:  res.Metadata.Timestamp,
			},
			PHash: res.PHash,
		})
	}

	// Generate reports using existing report package
	templatePath := "web/templates/dashboard.html"
	htmlPath := filepath.Join(a.outputPath, "report.html")
	if err := report.GenerateHTML(legacyResults, templatePath, htmlPath); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	return nil
}
