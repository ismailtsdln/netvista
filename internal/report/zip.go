package report

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func CreateReportZip(reportDir string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		relPath, err := filepath.Rel(reportDir, path)
		if err != nil {
			return err
		}

		w, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, f)
		return err
	})

	return err
}
