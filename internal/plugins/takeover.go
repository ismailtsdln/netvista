package plugins

import (
	"github.com/ismailtsdln/netvista/pkg/models"
)

type TakeoverPlugin struct {
	Signatures map[string]string
}

func NewTakeoverPlugin() *TakeoverPlugin {
	return &TakeoverPlugin{
		Signatures: map[string]string{
			"GitHub Pages": "There isn't a GitHub Pages site here",
			"Heroku":       "herokucdn.com/error-pages/no-such-app.html",
			"S3 Bucket":    "The specified bucket does not exist",
		},
	}
}

func (t *TakeoverPlugin) Name() string {
	return "Domain Takeover Detection"
}

func (t *TakeoverPlugin) Execute(target *models.Target) error {
	// This would ideally check the body, but for now we'll just check if it's already in metadata or placeholder
	// In a real scenario, we'd pass the body to Execute or re-fetch (not ideal)
	return nil
}
