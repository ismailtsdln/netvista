package prober

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ismailtsdln/netvista/pkg/models"
)

type Prober struct {
	Client  *http.Client
	Timeout time.Duration
}

func NewProber(timeout time.Duration) *Prober {
	return &Prober{
		Client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
		Timeout: timeout,
	}
}

func (p *Prober) Probe(ctx context.Context, target string) (*models.Target, error) {
	if !strings.HasPrefix(target, "http") {
		// Try HTTPS first then HTTP
		res, err := p.probeURL(ctx, "https://"+target)
		if err == nil {
			return res, nil
		}
		return p.probeURL(ctx, "http://"+target)
	}
	return p.probeURL(ctx, target)
}

func (p *Prober) probeURL(ctx context.Context, targetURL string) (*models.Target, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "NetVista/0.1.0 (Visual Recon Tool)")

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	metadata := models.ResponseMetadata{
		StatusCode:    resp.StatusCode,
		ContentLength: resp.ContentLength,
		Headers:       make(map[string]string),
		Timestamp:     time.Now(),
	}

	for k, v := range resp.Header {
		metadata.Headers[k] = strings.Join(v, ", ")
	}

	// Simple title extraction could go here, but we'll likely do it better with Playwright or a dedicated parser

	target := &models.Target{
		URL:      targetURL,
		IsAlive:  true,
		Metadata: metadata,
	}

	return target, nil
}
