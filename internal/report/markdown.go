package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/ismailtsdln/netvista/pkg/models"
)

// ExportMarkdown writes the scan results to a Markdown file.
func ExportMarkdown(results []models.Target, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "# NetVista Scan Report")
	fmt.Fprintf(file, "\nGenerated on: %s\n\n", results[0].Metadata.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintln(file, "| URL | Status | Title | Technology |")
	fmt.Fprintln(file, "| :--- | :--- | :--- | :--- |")

	for _, r := range results {
		fmt.Fprintf(file, "| %s | %d | %s | %s |\n",
			r.URL,
			r.Metadata.StatusCode,
			r.Metadata.Title,
			strings.Join(r.Metadata.Technology, ", "),
		)
	}

	return nil
}
