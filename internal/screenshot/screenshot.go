package screenshot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/corona10/goimagehash"
	"github.com/playwright-community/playwright-go"
)

type CaptureResult struct {
	Path  string
	PHash string
}

type Capturer struct {
	PW        *playwright.Playwright
	Browser   playwright.Browser
	OutputDir string
	ProxyURL  string
}

func NewCapturer(outputDir string, proxyURL string) (*Capturer, error) {
	err := playwright.Install()
	if err != nil {
		return nil, fmt.Errorf("could not install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}

	return &Capturer{
		PW:        pw,
		Browser:   browser,
		OutputDir: outputDir,
		ProxyURL:  proxyURL,
	}, nil
}

func (c *Capturer) Capture(ctx context.Context, url string, filename string) (*CaptureResult, error) {
	opts := playwright.BrowserNewContextOptions{}
	if c.ProxyURL != "" {
		opts.Proxy = &playwright.Proxy{
			Server: c.ProxyURL,
		}
	}

	context, err := c.Browser.NewContext(opts)
	if err != nil {
		return nil, err
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		return nil, err
	}

	// Smart Rendering: Use NetworkIdle for JS-heavy apps
	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return nil, err
	}

	path := filepath.Join(c.OutputDir, filename)
	screenshotBytes, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
	})
	if err != nil {
		return nil, err
	}

	// Generate PHash
	phash := ""
	img, _, err := image.Decode(bytes.NewReader(screenshotBytes))
	if err == nil {
		hash, err := goimagehash.PerceptionHash(img)
		if err == nil {
			phash = hash.ToString()
		}
	}

	return &CaptureResult{
		Path:  path,
		PHash: phash,
	}, nil
}

func (c *Capturer) Close() error {
	var err error
	if c.Browser != nil {
		err = c.Browser.Close()
	}
	if c.PW != nil {
		stopErr := c.PW.Stop()
		if err == nil {
			err = stopErr
		}
	}
	return err
}
