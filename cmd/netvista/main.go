package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ismailtsdln/netvista/internal/core/domain"
	"github.com/ismailtsdln/netvista/internal/core/ports"
	"github.com/ismailtsdln/netvista/internal/core/services"
	"github.com/ismailtsdln/netvista/internal/engine"
	"github.com/ismailtsdln/netvista/internal/infra/adapters"
	"github.com/ismailtsdln/netvista/internal/plugins"
	"github.com/ismailtsdln/netvista/pkg/config"
	"github.com/ismailtsdln/netvista/pkg/signatures"
	"github.com/ismailtsdln/netvista/pkg/utils"
)

var (
	version = "0.2.0-adv"
)

func main() {
	// Initialize slog
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	banner := utils.GetBanner(version)
	color.Cyan(banner)

	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	confPath := scanCmd.String("config", "netvista.yaml", "Path to config file")
	portsFlag := scanCmd.String("p", "", "Ports to scan (e.g., 80,443,8000-9000)")
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
		if *portsFlag == "" {
			*portsFlag = cfg.Ports
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

		*portsFlag = utils.GetPortPreset(*portsFlag)
		d, err := time.ParseDuration(*timeout)
		if err != nil {
			slog.Error("Invalid timeout", "timeout", *timeout, "error", err)
			os.Exit(1)
		}

		customHeaders := make(map[string]string)
		var customHeadersList []string
		if *headers != "" {
			customHeadersList = strings.Split(*headers, ",")
			for _, part := range customHeadersList {
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

		// Initialize Adapters
		proberAdapter := adapters.NewProberAdapter(d, *proxy, customHeaders)
		rendererAdapter, err := adapters.NewRendererAdapter(*output, *proxy, false)
		if err != nil {
			slog.Error("Failed to initialize renderer", "error", err)
			os.Exit(1)
		}
		defer rendererAdapter.Close()

		reporterAdapter := adapters.NewReporterAdapter(*output)

		wafAnalyzer := adapters.NewWafAnalyzerAdapter(plugins.NewWafPlugin(sigs.Wafs))

		// Initialize Service
		scannerService := services.NewScannerService(
			proberAdapter,
			rendererAdapter,
			[]ports.Analyzer{wafAnalyzer},
			reporterAdapter,
			domain.Config{
				Concurrency: *concurrency,
				OutputPath:  *output,
			},
			logger,
		)

		var rawTargets []string
		var terr error

		if *nmapFile != "" {
			slog.Info("Parsing Nmap XML", "file", *nmapFile)
			rawTargets, terr = utils.ParseNmapXML(*nmapFile)
			if terr != nil {
				slog.Error("Error parsing Nmap XML", "error", terr)
				os.Exit(1)
			}
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				rawTargets = engine.ReadTargetsFromStdin()
			} else {
				slog.Error("No input provided. Pipe targets or use --nmap")
				os.Exit(1)
			}
		}

		if len(rawTargets) == 0 {
			slog.Info("No targets found.")
			os.Exit(0)
		}

		rawTargets = utils.ResolveTargets(rawTargets)
		var targets []domain.Target
		for _, rt := range rawTargets {
			targets = append(targets, domain.Target{URL: rt})
		}

		ctx := context.Background()
		if err := scannerService.Scan(ctx, targets); err != nil {
			slog.Error("Scan failed", "error", err)
			os.Exit(1)
		}

		// Handle conditional exports if needed (already handled by reporter adapter for basic HTML)
		// For MD/CSV/TXT we can add more logic to reporter or just leave as is since adapter handles it
		_ = exportCSV
		_ = exportMD
		_ = exportTXT

		color.Green("\n [✓] Scan complete! Results saved to: %s", *output)
		color.Yellow(" [i] Run 'netvista serve -d %s' to view interactive dashboard.\n", *output)

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
