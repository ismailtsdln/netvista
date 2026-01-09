package report

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/ismailtsdln/netvista/pkg/models"
)

type ReportData struct {
	Results []models.Target
}

func GenerateHTML(results []models.Target, templatePath string, outputPath string) error {
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
