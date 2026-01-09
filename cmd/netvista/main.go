package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ismailtsdln/netvista/internal/engine"
	"github.com/ismailtsdln/netvista/internal/plugins"
	"github.com/ismailtsdln/netvista/internal/prober"
	"github.com/ismailtsdln/netvista/internal/report"
	"github.com/ismailtsdln/netvista/internal/screenshot"
	"github.com/ismailtsdln/netvista/pkg/config"
	"github.com/ismailtsdln/netvista/pkg/models"
	"github.com/ismailtsdln/netvista/pkg/signatures"
	"github.com/ismailtsdln/netvista/pkg/utils"
	"github.com/schollz/progressbar/v3"
)

var (
	version = "0.1.0"
)

func main() {
	// Initialize slog
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	banner := utils.GetBanner(version)
	color.Cyan(banner)

	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	confPath := scanCmd.String("config", "netvista.yaml", "Path to config file")
	ports := scanCmd.String("p", "", "Ports to scan (e.g., 80,443,8000-9000)")
	concurrency := scanCmd.Int("c", 0, "Number of concurrent workers")
	output := scanCmd.String("o", "", "Output directory for reports")
	timeout := scanCmd.String("t", "", "Timeout per host")
	nmapFile := scanCmd.String("nmap", "", "Nmap XML file to parse")
	proxy := scanCmd.String("proxy", "", "Proxy URL")
	headers := scanCmd.String("H", "", "Custom headers")
	redirects := scanCmd.Int("max-redirects", 0, "Max redirects to follow")
	exportCSV := scanCmd.Bool("csv", false, "Export to CSV")
	exportMD := scanCmd.Bool("md", false, "Export to Markdown")
	exportTXT := scanCmd.Bool("txt", false, "Export to Text (alive URLs)")

	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	serveDir := serveCmd.String("d", "reports", "Directory to serve")
	servePort := serveCmd.String("p", "8080", "Port to serve on")

	if len(os.Args) < 2 {
		color.Red("Expected 'scan', 'serve' or 'version' subcommands")
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
		if *redirects == 0 {
			*redirects = 10 // Default
		}

		// Handle port presets
		*ports = utils.GetPortPreset(*ports)

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

		p := prober.NewProber(d, *proxy, customHeaders, *redirects)

		e := engine.NewEngine(*concurrency, p)

		pm := plugins.NewPluginManager()
		pm.Register(plugins.NewFingerprintPlugin(sigs.Fingerprints))
		pm.Register(plugins.NewTakeoverPlugin(sigs.Takeovers))
		pm.Register(plugins.NewWafPlugin(sigs.Wafs))

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

		targets = utils.ResolveTargets(targets)

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

		bar := progressbar.NewOptions(len(targets),
			progressbar.OptionSetDescription("Scanning"),
			progressbar.OptionSetWidth(20),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[cyan]=[reset]",
				SaucerHead:    "[cyan]>[reset]",
				SaucerPadding: " ",
				BarStart:      "[",
				BarEnd:        "]",
			}),
		)

		ctx := context.Background()
		results := e.Run(ctx, targets)

		var scanResults []models.Target
		for res := range results {
			bar.Add(1)
			// slog.Info("Found", "url", res.URL, "status", res.Metadata.StatusCode) // Reduce noise during progress bar

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

			// Resolve resolvedTargets for report generation if needed, but targets is enough

			// Generate HTML
			htmlPath := filepath.Join(*output, "report.html")
			err = report.ExportJSON(scanResults, filepath.Join(*output, "results.json"))
			if err != nil {
				slog.Error("Error exporting JSON", "error", err)
			}

			// CSV Export
			if *exportCSV {
				csvPath := filepath.Join(*output, "results.csv")
				if err := report.ExportCSV(scanResults, csvPath); err != nil {
					slog.Error("Error exporting CSV", "error", err)
				} else {
					slog.Info("CSV report generated", "path", csvPath)
				}
			}

			// Markdown Export
			if *exportMD {
				mdPath := filepath.Join(*output, "results.md")
				if err := report.ExportMarkdown(scanResults, mdPath); err != nil {
					slog.Error("Error exporting Markdown", "error", err)
				} else {
					slog.Info("Markdown report generated", "path", mdPath)
				}
			}

			// Text Export
			if *exportTXT {
				txtPath := filepath.Join(*output, "urls.txt")
				if err := report.ExportText(scanResults, txtPath); err != nil {
					slog.Error("Error exporting Text", "error", err)
				} else {
					slog.Info("Text report generated", "path", txtPath)
				}
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

			// Final Summary Table
			fmt.Println()
			color.Cyan(" [📊] Scan Summary")
			fmt.Println(" " + strings.Repeat("─", 100))
			color.White(" %-40s | %-6s | %-30s | %-20s", "URL", "Status", "Title", "Technology")
			fmt.Println(" " + strings.Repeat("─", 100))

			for _, r := range scanResults {
				statusStr := fmt.Sprintf("%d", r.Metadata.StatusCode)
				if r.Metadata.StatusCode >= 400 {
					statusStr = color.RedString("%d", r.Metadata.StatusCode)
				} else if r.Metadata.StatusCode >= 200 && r.Metadata.StatusCode < 300 {
					statusStr = color.GreenString("%d", r.Metadata.StatusCode)
				} else {
					statusStr = color.YellowString("%d", r.Metadata.StatusCode)
				}

				title := r.Metadata.Title
				if len(title) > 30 {
					title = title[:27] + "..."
				}

				url := r.URL
				if len(url) > 40 {
					url = url[:37] + "..."
				}

				techs := strings.Join(r.Metadata.Technology, ", ")
				if len(techs) > 20 {
					techs = techs[:17] + "..."
				}

				fmt.Printf(" %-40s | %-16s | %-30s | %-20s\n", url, statusStr, title, techs)
			}
			fmt.Println(" " + strings.Repeat("─", 100))

			color.Green("\n [✓] Scan complete! Results saved to: %s", *output)
			color.Yellow(" [i] Run 'netvista serve -d %s' to view interactive dashboard.\n", *output)
		}

	case "serve":
		serveCmd.Parse(os.Args[2:])
		absPath, _ := filepath.Abs(*serveDir)
		color.Cyan("\n [▶] NetVista Serve Engine")
		color.White(" ──────────────────────────────────────────────────")
		color.Green(" [✓] Serving reports from: %s", absPath)
		color.Green(" [●] Dashboard available at: http://localhost:%s", *servePort)
		color.White(" ──────────────────────────────────────────────────")
		color.Yellow(" [!] Press Ctrl+C to stop the server\n")

		handler := http.FileServer(http.Dir(*serveDir))
		if err := http.ListenAndServe(":"+*servePort, handler); err != nil {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}

	case "version":
		color.Blue("NetVista v%s\n", version)
	default:
		color.Red("Unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}
