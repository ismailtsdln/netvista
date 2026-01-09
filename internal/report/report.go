package report

import (
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ismailtsdln/netvista/internal/screenshot"
	"github.com/ismailtsdln/netvista/pkg/models"
	"github.com/ismailtsdln/netvista/web"
)

type ReportData struct {
	Results []models.Target
	Grouped map[string][]models.Target
}

func GenerateHTML(results []models.Target, templatePath string, outputPath string) error {
	// Ensure assets directory exists in output
	outDir := filepath.Dir(outputPath)
	assetsDir := filepath.Join(outDir, "assets")
	os.MkdirAll(assetsDir, 0755)

	// Copy logo
	logoDst := filepath.Join(assetsDir, "logo.png")

	// Try to get logo from embedded first as it's more reliable for installed binaries
	logoData, err := fs.ReadFile(web.AssetsFS, "assets/logo.png")
	if err == nil {
		os.WriteFile(logoDst, logoData, 0644)
	} else {
		// Fallback to local file
		logoSrc := filepath.Join("web", "assets", "logo.png")
		if _, err := os.Stat(logoSrc); err == nil {
			s, err := os.Open(logoSrc)
			if err == nil {
				defer s.Close()
				d, err := os.Create(logoDst)
				if err == nil {
					defer d.Close()
					io.Copy(d, s)
				}
			}
		}
	}

	start := time.Now()
	// ...

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
			if err == nil && dist < 12 { // Increased threshold slightly for better grouping
				grouped[groupKey] = append(grouped[groupKey], r2)
				processed[r2.URL] = true
			}
		}
	}
	slog.Info("Clustering complete", "duration", time.Since(start), "groups", len(grouped))

	var tmpl *template.Template
	var terr error

	funcMap := template.FuncMap{
		"sanitize": func(s string) string {
			s = strings.Replace(s, "https://", "", 1)
			s = strings.Replace(s, "http://", "", 1)
			s = strings.ReplaceAll(s, "/", "_")
			s = strings.ReplaceAll(s, ":", "_")
			s = strings.ReplaceAll(s, ".", "_")
			return strings.Trim(s, "_")
		},
	}

	// Try local template first
	if _, err := os.Stat(templatePath); err == nil {
		tmpl, terr = template.New(filepath.Base(templatePath)).Funcs(funcMap).ParseFiles(templatePath)
	} else {
		// Fallback to embedded
		embeddedPath := filepath.Join("templates", filepath.Base(templatePath))
		tmpl, terr = template.New(filepath.Base(templatePath)).Funcs(funcMap).ParseFS(web.AssetsFS, embeddedPath)
	}

	if terr != nil {
		return terr
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
