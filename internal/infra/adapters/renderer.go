package adapters

import (
	"context"
	"fmt"

	"github.com/ismailtsdln/netvista/internal/core/domain"
	"github.com/ismailtsdln/netvista/internal/screenshot"
	"github.com/ismailtsdln/netvista/pkg/utils"
)

// RendererAdapter wraps the existing screenshot logic.
type RendererAdapter struct {
	c *screenshot.Capturer
}

// NewRendererAdapter creates a new renderer adapter.
func NewRendererAdapter(outputPath string, proxy string, fullPage bool) (*RendererAdapter, error) {
	c, err := screenshot.NewCapturer(outputPath, proxy)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize capturer: %w", err)
	}
	return &RendererAdapter{c: c}, nil
}

// Render captures a screenshot and returns the path and pHash.
func (a *RendererAdapter) Render(ctx context.Context, target domain.Target) (string, string, error) {
	// Clean filename for filesystem compatibility
	filename := fmt.Sprintf("%s.png", utils.SanitizeFilename(target.URL))
	res, err := a.c.Capture(ctx, target.URL, filename)
	if err != nil {
		return "", "", err
	}
	return res.Path, res.PHash, nil
}

// Close releases browser resources.
func (a *RendererAdapter) Close() error {
	return a.c.Close()
}
