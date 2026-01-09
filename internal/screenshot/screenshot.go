package screenshot

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/playwright-community/playwright-go"
)

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

func (c *Capturer) Capture(url string, filename string) ([]byte, error) {
	opts := playwright.BrowserNewContextOptions{}
	if c.ProxyURL != "" {
		opts.Proxy = &playwright.Proxy{
			Server: playwright.String(c.ProxyURL),
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

	if _, err = page.Goto(url); err != nil {
		return nil, err
	}

	// Wait for network to be idle
	page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})

	path := filepath.Join(c.OutputDir, filename)
	screenshotBytes, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
	})
	if err != nil {
		return nil, err
	}

	return screenshotBytes, nil
}

func (c *Capturer) Close() {
	if c.Browser != nil {
		c.Browser.Close()
	}
	if c.PW != nil {
		c.PW.Stop()
	}
}
