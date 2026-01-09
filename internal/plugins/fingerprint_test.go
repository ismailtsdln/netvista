package plugins

import (
	"testing"

	"github.com/ismailtsdln/netvista/pkg/models"
)

func TestFingerprintPlugin(t *testing.T) {
	plugin := NewFingerprintPlugin()

	target := &models.Target{
		Metadata: models.ResponseMetadata{
			Title: "WordPress Website",
			Headers: map[string]string{
				"Server": "nginx",
			},
		},
	}

	err := plugin.Execute(target)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	foundWP := false
	foundNginx := false
	for _, tech := range target.Metadata.Technology {
		if tech == "WordPress" {
			foundWP = true
		}
		if tech == "nginx" {
			foundNginx = true
		}
	}

	if !foundWP {
		t.Error("WordPress not detected in title")
	}
	if !foundNginx {
		t.Error("nginx not detected in headers")
	}
}
