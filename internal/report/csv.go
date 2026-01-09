package report

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"

	"github.com/ismailtsdln/netvista/pkg/models"
)

// ExportCSV writes the scan results to a CSV file.
func ExportCSV(results []models.Target, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"URL", "Status", "Title", "Technology", "PHash", "Timestamp"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, r := range results {
		row := []string{
			r.URL,
			strconv.Itoa(r.Metadata.StatusCode),
			r.Metadata.Title,
			strings.Join(r.Metadata.Technology, ", "),
			r.PHash,
			r.Metadata.Timestamp.String(),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
