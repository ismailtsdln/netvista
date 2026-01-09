package report

import (
	"fmt"
	"os"

	"github.com/ismailtsdln/netvista/pkg/models"
)

// ExportText writes a clean list of alive URLs to a text file.
func ExportText(results []models.Target, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, r := range results {
		fmt.Fprintln(file, r.URL)
	}

	return nil
}
