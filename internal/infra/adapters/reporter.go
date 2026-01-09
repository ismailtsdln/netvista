package adapters

import (
	"context"
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
	return &ReporterAdapter{outputPath: outputPath}
}

// Report generates reports based on scan results.
func (a *ReporterAdapter) Report(ctx context.Context, results []domain.ScanResult) error {
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

	// 1. JSON Export
	jsonPath := filepath.Join(a.outputPath, "results.json")
	if err := report.ExportJSON(legacyResults, jsonPath); err != nil {
		return fmt.Errorf("failed to export JSON: %w", err)
	}

	// 2. CSV Export
	csvPath := filepath.Join(a.outputPath, "results.csv")
	report.ExportCSV(legacyResults, csvPath) // Non-blocking/Best-effort

	// 3. Markdown Export
	mdPath := filepath.Join(a.outputPath, "results.md")
	report.ExportMarkdown(legacyResults, mdPath)

	// 4. Text Export
	txtPath := filepath.Join(a.outputPath, "urls.txt")
	report.ExportText(legacyResults, txtPath)

	// 5. HTML Export
	templatePath := "web/templates/dashboard.html"
	htmlPath := filepath.Join(a.outputPath, "report.html")
	if err := report.GenerateHTML(legacyResults, templatePath, htmlPath); err != nil {
		return fmt.Errorf("failed to generate HTML report: %w", err)
	}

	// 6. ZIP Export
	zipPath := filepath.Join(a.outputPath, "report.zip")
	report.CreateReportZip(a.outputPath, zipPath)

	return nil
}
