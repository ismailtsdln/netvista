package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ismailtsdln/netvista/internal/engine"
	"github.com/ismailtsdln/netvista/internal/plugins"
	"github.com/ismailtsdln/netvista/internal/prober"
	"github.com/ismailtsdln/netvista/internal/report"
	"github.com/ismailtsdln/netvista/internal/screenshot"
	"github.com/ismailtsdln/netvista/pkg/config"
	"github.com/ismailtsdln/netvista/pkg/models"
	"github.com/ismailtsdln/netvista/pkg/signatures"
	"github.com/ismailtsdln/netvista/pkg/utils"
)

var (
	version = "0.1.0"
)

func main() {
	// Initialize slog
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	banner := utils.GetBanner(version)
	fmt.Println(banner)

	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	confPath := scanCmd.String("config", "netvista.yaml", "Path to config file")
	ports := scanCmd.String("p", "", "Ports to scan (e.g., 80,443,8000-9000)")
	concurrency := scanCmd.Int("c", 0, "Number of concurrent workers")
	output := scanCmd.String("o", "", "Output directory for reports")
	timeout := scanCmd.String("t", "", "Timeout per host")
	nmapFile := scanCmd.String("nmap", "", "Nmap XML file to parse")
	proxy := scanCmd.String("proxy", "", "Proxy URL")
	headers := scanCmd.String("H", "", "Custom headers")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'scan' subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "scan":
		scanCmd.Parse(os.Args[2:])

		// Load YAML config
		cfg, err := config.LoadConfig(*confPath)
		if err != nil {
			slog.Error("Failed to load config", "error", err)
			os.Exit(1)
		}

		// Merge Flags (CLI overrides YAML)
		if *ports == "" {
			*ports = cfg.Ports
		}
		if *concurrency == 0 {
			*concurrency = cfg.Concurrency
		}
		if *output == "" {
			*output = cfg.Output
		}
		if *timeout == "" {
			*timeout = cfg.Timeout
		}
		if *proxy == "" {
			*proxy = cfg.Proxy
		}
		if *headers == "" {
			*headers = cfg.Headers
		}

		d, err := time.ParseDuration(*timeout)
		if err != nil {
			slog.Error("Invalid timeout", "timeout", *timeout, "error", err)
			os.Exit(1)
		}

		// Parse headers

		// Parse headers
		customHeaders := make(map[string]string)
		if *headers != "" {
			parts := strings.Split(*headers, ",")
			for _, part := range parts {
				kv := strings.SplitN(part, ":", 2)
				if len(kv) == 2 {
					customHeaders[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}
			}
		}

		// Load Signatures
		sigPath := filepath.Join("pkg", "signatures", "signatures.yaml")
		sigs, err := signatures.LoadSignatures(sigPath)
		if err != nil {
			slog.Warn("Failed to load signatures, some features may be limited", "path", sigPath, "error", err)
			sigs = &signatures.Signatures{}
		}

		p := prober.NewProber(d, *proxy, customHeaders)
		e := engine.NewEngine(*concurrency, p)

		pm := plugins.NewPluginManager()
		pm.Register(plugins.NewFingerprintPlugin(sigs.Fingerprints))
		pm.Register(plugins.NewTakeoverPlugin(sigs.Takeovers))

		var targets []string
		var terr error

		if *nmapFile != "" {
			slog.Info("Parsing Nmap XML", "file", *nmapFile)
			targets, terr = utils.ParseNmapXML(*nmapFile)
			if terr != nil {
				slog.Error("Error parsing Nmap XML", "error", terr)
				os.Exit(1)
			}
		} else {
			// Check if piped input
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				targets = engine.ReadTargetsFromStdin()
			} else {
				slog.Error("No input provided. Pipe targets or use --nmap")
				os.Exit(1)
			}
		}

		if len(targets) == 0 {
			slog.Info("No targets found.")
			os.Exit(0)
		}

		slog.Info("Starting scan",
			"targets", len(targets),
			"ports", *ports,
			"concurrency", *concurrency,
			"output", *output)

		cap, err := screenshot.NewCapturer(*output, *proxy)
		if err != nil {
			slog.Error("Error initializing screenshot engine", "error", err)
			os.Exit(1)
		}
		defer cap.Close()

		ctx := context.Background()
		results := e.Run(ctx, targets)

		var scanResults []models.Target
		for res := range results {
			slog.Info("Found", "url", res.URL, "status", res.Metadata.StatusCode)

			// Run plugins
			pm.RunAll(res)
			if len(res.Metadata.Technology) > 0 {
				slog.Info("Technologies", "url", res.URL, "tech", strings.Join(res.Metadata.Technology, ", "))
			}

			// Hostname for filename
			filename := fmt.Sprintf("%s.png", utils.SanitizeFilename(res.URL))
			imgBytes, err := cap.Capture(res.URL, filename)
			if err != nil {
				slog.Error("Error capturing screenshot", "url", res.URL, "error", err)
			} else {
				slog.Info("Screenshot saved", "url", res.URL, "filename", filename)

				// Generate PHash
				phash, err := screenshot.GeneratePHash(imgBytes)
				if err == nil {
					res.PHash = phash
					slog.Info("PHash generated", "url", res.URL, "phash", phash)
				}
			}
			scanResults = append(scanResults, *res)
		}

		if len(scanResults) > 0 {
			slog.Info("Generating reports...")

			// Generate HTML
			htmlPath := filepath.Join(*output, "report.html")
			// We need a way to pass sanitize function to template
			// For simplicity in this implementation, I'll update report.go or main.go to use a pre-sanitized name if needed
			// Let's update models and report logic to handle this better

			err = report.ExportJSON(scanResults, filepath.Join(*output, "results.json"))
			if err != nil {
				slog.Error("Error exporting JSON", "error", err)
			}

			// For HTML we need the template. We'll use a relative path for now.
			templatePath := "web/templates/dashboard.html"
			err = report.GenerateHTML(scanResults, templatePath, htmlPath)
			if err != nil {
				slog.Error("Error generating HTML report", "error", err)
			} else {
				slog.Info("HTML report generated", "path", htmlPath)
			}

			// Generate ZIP report
			zipPath := filepath.Join(*output, "report.zip")
			if err := report.CreateReportZip(*output, zipPath); err != nil {
				slog.Error("Error generating ZIP", "error", err)
			} else {
				slog.Info("ZIP report generated", "path", zipPath)
			}
		}

	case "version":
		fmt.Printf("NetVista v%s\n", version)
	default:
		fmt.Println("Expected 'scan' or 'version' subcommands")
		os.Exit(1)
	}
}
