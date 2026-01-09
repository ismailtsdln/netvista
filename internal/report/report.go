package report

import (
	"encoding/json"
	"html/template"
	"os"

	"github.com/ismailtsdln/netvista/pkg/models"
)

type ReportData struct {
	Results []models.Target
}

func GenerateHTML(results []models.Target, templatePath string, outputPath string) error {
	tmpl, err := template.ParseFiles(templatePath)
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
