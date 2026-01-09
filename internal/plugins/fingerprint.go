package plugins

import (
	"strings"

	"github.com/ismailtsdln/netvista/pkg/models"
)

type FingerprintPlugin struct{}

func (f *FingerprintPlugin) Name() string {
	return "Technology Fingerprinting"
}

func (f *FingerprintPlugin) Execute(target *models.Target) error {
	// Simple header-based detection
	server := target.Metadata.Headers["Server"]
	if server != "" {
		target.Metadata.Technology = append(target.Metadata.Technology, server)
	}

	poweredBy := target.Metadata.Headers["X-Powered-By"]
	if poweredBy != "" {
		target.Metadata.Technology = append(target.Metadata.Technology, poweredBy)
	}

	// Simple title-based (placeholder for more complex logic)
	if strings.Contains(target.Metadata.Title, "WordPress") {
		target.Metadata.Technology = append(target.Metadata.Technology, "WordPress")
	}

	return nil
}
