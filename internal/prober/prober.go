package prober

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ismailtsdln/netvista/pkg/models"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/119.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
}

type Prober struct {
	Client        *http.Client
	Timeout       time.Duration
	ProxyURL      string
	CustomHeaders map[string]string
	Redirects     int
}

func NewProber(timeout time.Duration, proxyURL string, customHeaders map[string]string, redirects int) *Prober {
	transport := &http.Transport{}
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxy)
		}
	}

	return &Prober{
		Client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= redirects {
					return fmt.Errorf("stopped after %d redirects", redirects)
				}
				return nil
			},
		},
		Timeout:       timeout,
		ProxyURL:      proxyURL,
		CustomHeaders: customHeaders,
		Redirects:     redirects,
	}
}

func (p *Prober) getRandomUA() string {
	return userAgents[rand.Intn(len(userAgents))]
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

	req.Header.Set("User-Agent", p.getRandomUA())

	for k, v := range p.CustomHeaders {
		req.Header.Set(k, v)
	}

	resp, err := p.Client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Limit body read to 1MB
	const maxBodySize = 1 * 1024 * 1024 // 1MB
	limitedReader := io.LimitReader(resp.Body, maxBodySize)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		// Log error but continue, as we might still have headers/status
		fmt.Printf("Error reading response body for %s: %v\n", targetURL, err)
	}
	body := string(bodyBytes)

	// Simple title extraction
	titleRegex := regexp.MustCompile("(?i)<title>(.*?)</title>")

	title := "No Title"
	matches := titleRegex.FindStringSubmatch(body)
	if len(matches) > 1 {
		title = matches[1]
	}

	metadata := models.ResponseMetadata{
		Title:      title,
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Body:       body,
		Timestamp:  time.Now(),
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
