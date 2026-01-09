package report

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/ismailtsdln/netvista/internal/screenshot"
	"github.com/ismailtsdln/netvista/pkg/models"
)

type ReportData struct {
	Results []models.Target
	Grouped map[string][]models.Target
}

func GenerateHTML(results []models.Target, templatePath string, outputPath string) error {
	// Simple clustering
	grouped := make(map[string][]models.Target)
	processed := make(map[string]bool)

	for i, r1 := range results {
		if processed[r1.URL] {
			continue
		}

		groupKey := r1.URL
		grouped[groupKey] = append(grouped[groupKey], r1)
		processed[r1.URL] = true

		if r1.PHash == "" {
			continue
		}

		for j := i + 1; j < len(results); j++ {
			r2 := results[j]
			if processed[r2.URL] || r2.PHash == "" {
				continue
			}

			dist, err := screenshot.HammingDistance(r1.PHash, r2.PHash)
			if err == nil && dist < 8 { // Threshold for similarity
				grouped[groupKey] = append(grouped[groupKey], r2)
				processed[r2.URL] = true
			}
		}
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(template.FuncMap{
		"sanitize": func(s string) string {
			// Basic sanitization similar to utils.SanitizeFilename
			s = strings.Replace(s, "https://", "", 1)
			s = strings.Replace(s, "http://", "", 1)
			s = strings.ReplaceAll(s, "/", "_")
			s = strings.ReplaceAll(s, ":", "_")
			s = strings.ReplaceAll(s, ".", "_")
			return strings.Trim(s, "_")
		},
	}).ParseFiles(templatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	data := ReportData{
		Results: results,
		Grouped: grouped,
	}

	return tmpl.Execute(f, data)
}

func ExportJSON(results []models.Target, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}
