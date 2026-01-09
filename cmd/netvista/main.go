package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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
	"github.com/ismailtsdln/netvista/pkg/models"
	"github.com/ismailtsdln/netvista/pkg/signatures"
	"github.com/ismailtsdln/netvista/pkg/utils"
)

var (
	version = "2.1.0-pro"
)

func main() {
	// Initialize slog
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	banner := utils.GetBanner(version)
	color.Cyan(banner)

	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	scanCmd.Usage = func() {
		color.Cyan("\n [▶] NetVista Scan Engine - Usage")
		fmt.Fprintf(os.Stderr, " Usage: netvista scan [options] < targets.txt\n\n")
		color.Yellow(" Options:")
		scanCmd.PrintDefaults()
		color.Yellow("\n Examples:")
		fmt.Println("  # Basic scanning from stdin")
		fmt.Println("  echo \"example.com\" | ./netvista scan -o reports")
		fmt.Println("\n  # Comprehensive scan with custom ports and concurrency")
		fmt.Println("  cat targets.txt | ./netvista scan -p 80,443,8080 -c 10 -o results")
		fmt.Println("\n  # Resuming a scan (Incremental mode)")
		fmt.Println("  echo \"example.com\" | ./netvista scan -o previous_results")
		fmt.Println("\n  # Parsing Nmap XML input")
		fmt.Println("  ./netvista scan --nmap network.xml -o nmap_report")
		fmt.Println("\n  # Scan via Proxy and auto-open report")
		fmt.Println("  # Scan via Proxy and auto-open report")
		fmt.Println("  echo \"target.local\" | ./netvista scan -proxy \"http://127.0.0.1:8080\" -open")
	}

	confPath := scanCmd.String("config", "netvista.yaml", "Path to config file")
	portsFlag := scanCmd.String("p", "", "Ports to scan (preset: top-100, top-1000, full or e.g., 80,443)")
	concurrency := scanCmd.Int("c", 0, "Number of concurrent workers")
	output := scanCmd.String("o", "reports", "Output directory for reports")
	timeout := scanCmd.String("t", "10s", "Timeout per host")
	nmapFile := scanCmd.String("nmap", "", "Nmap XML file to parse")
	proxy := scanCmd.String("proxy", "", "Proxy URL (e.g., http://127.0.0.1:8080)")
	headers := scanCmd.String("H", "", "Custom headers (e.g., \"User-Agent: NetVista, X-Scan: 1\")")
	redirects := scanCmd.Int("max-redirects", 10, "Max redirects to follow")
	exportCSV := scanCmd.Bool("csv", true, "Export to CSV")
	exportMD := scanCmd.Bool("md", true, "Export to Markdown")
	exportTXT := scanCmd.Bool("txt", true, "Export to Text (alive URLs)")
	autoOpen := scanCmd.Bool("open", false, "Automatically open the HTML report")

	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	serveCmd.Usage = func() {
		color.Cyan("\n [▶] NetVista Serve Engine - Usage")
		fmt.Fprintf(os.Stderr, " Usage: netvista serve -d <report_dir> -p <port>\n\n")
		color.Yellow(" Options:")
		serveCmd.PrintDefaults()
		color.Yellow("\n Examples:")
		fmt.Println("  # Serve current reports directory on default port")
		fmt.Println("  ./netvista serve -d reports")
		fmt.Println("\n  # Serve custom directory on port 9090")
		fmt.Println("  # Serve custom directory on port 9090")
		fmt.Println("  ./netvista serve -d my_scan -p 9090")
	}
	serveDir := serveCmd.String("d", "reports", "Directory to serve")
	servePort := serveCmd.String("p", "8080", "Port to serve on")

	flag.Usage = func() {
		color.Cyan(utils.GetBanner(version))
		fmt.Println(" Advanced Visual Reconnaissance Tool for Professionals")
		color.Yellow(" Usage:")
		fmt.Println("  netvista <command> [options]")
		color.Yellow(" Commands:")
		fmt.Println("  scan    Perform network discovery and visual capture")
		fmt.Println("  serve   Launch interactive dashboard from report directory")
		fmt.Println("  version Show application version info")
		color.Yellow("\n Use 'netvista <command> --help' for more information on a command.\n")
	}

	if len(os.Args) < 2 {
		flag.Usage()
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

		// Initialize Adapters
		proberAdapter := adapters.NewProberAdapter(d, *proxy, customHeaders)
		rendererAdapter, err := adapters.NewRendererAdapter(*output, *proxy, false, cfg.MaxBrowserContexts)
		if err != nil {
			slog.Error("Failed to initialize renderer", "error", err)
			os.Exit(1)
		}
		defer rendererAdapter.Close()

		reporterAdapter := adapters.NewReporterAdapter(*output, *exportCSV, *exportMD, *exportTXT)

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

		// Load existing results for incremental scanning
		seenURLs := make(map[string]bool)
		resultsPath := filepath.Join(*output, "results.json")
		if data, terr := os.ReadFile(resultsPath); terr == nil {
			var existing []models.Target
			if err := json.Unmarshal(data, &existing); err == nil {
				for _, r := range existing {
					seenURLs[r.URL] = true
				}
				slog.Info("Loaded existing results for incremental scan", "count", len(existing))
			}
		}

		resolvedTargets := utils.ResolveTargets(rawTargets)
		var finalTargets []domain.Target
		for _, rt := range resolvedTargets {
			if seenURLs[rt] {
				continue
			}
			finalTargets = append(finalTargets, domain.Target{URL: rt})
		}

		if len(finalTargets) == 0 {
			slog.Info("All targets already processed.")
			os.Exit(0)
		}

		ctx := context.Background()
		if err := scannerService.Scan(ctx, finalTargets); err != nil {
			slog.Error("Scan failed", "error", err)
			os.Exit(1)
		}

		if *autoOpen {
			slog.Info("Opening report...")
			htmlPath := filepath.Join(*output, "report.html")
			utils.OpenBrowser(htmlPath)
		}
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
	case "-h", "--help":
		flag.Usage()
	default:
		color.Red("Unknown subcommand: %s", os.Args[1])
		fmt.Println("\nRun 'netvista --help' for usage info.")
		os.Exit(1)
	}
}
