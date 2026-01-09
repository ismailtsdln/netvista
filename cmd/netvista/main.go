package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ismailtsdln/netvista/internal/engine"
	"github.com/ismailtsdln/netvista/internal/prober"
	"github.com/ismailtsdln/netvista/internal/screenshot"
	"github.com/ismailtsdln/netvista/pkg/utils"
)

var (
	version = "0.1.0"
)

func main() {
	banner := utils.GetBanner(version)
	fmt.Println(banner)

	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	ports := scanCmd.String("p", "80,443", "Ports to scan (e.g., 80,443,8000-9000)")
	concurrency := scanCmd.Int("c", 20, "Number of concurrent workers")
	output := scanCmd.String("o", "./out", "Output directory for reports")
	timeout := scanCmd.String("t", "5s", "Timeout per host")
	nmapFile := scanCmd.String("nmap", "", "Nmap XML file to parse")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'scan' subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "scan":
		scanCmd.Parse(os.Args[2:])

		d, err := time.ParseDuration(*timeout)
		if err != nil {
			fmt.Printf("Invalid timeout: %v\n", err)
			os.Exit(1)
		}

		p := prober.NewProber(d)
		e := engine.NewEngine(*concurrency, p)

		var targets []string
		if *nmapFile != "" {
			fmt.Printf("Parsing Nmap XML: %s\n", *nmapFile)
			targets, err = utils.ParseNmapXML(*nmapFile)
			if err != nil {
				fmt.Printf("Error parsing Nmap XML: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Check if piped input
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				targets = engine.ReadTargetsFromStdin()
			} else {
				fmt.Println("No input provided. Pipe targets or use --nmap")
				os.Exit(1)
			}
		}

		if len(targets) == 0 {
			fmt.Println("No targets found.")
			os.Exit(0)
		}

		// If ports are specified and it's not from nmap, we might want to expand targets
		// For now, we'll just log that we're using the specified ports if applicable
		fmt.Printf("Starting scan on %d targets with ports [%s], concurrency: %d, output: %s\n", len(targets), *ports, *concurrency, *output)

		cap, err := screenshot.NewCapturer(*output)
		if err != nil {
			fmt.Printf("Error initializing screenshot engine: %v\n", err)
			os.Exit(1)
		}
		defer cap.Close()

		ctx := context.Background()
		results := e.Run(ctx, targets)

		for res := range results {
			fmt.Printf("[+] Found: %s (%d)\n", res.URL, res.Metadata.StatusCode)

			// Hostname for filename
			filename := fmt.Sprintf("%s.png", utils.SanitizeFilename(res.URL))
			err := cap.Capture(res.URL, filename)
			if err != nil {
				fmt.Printf(" [!] Error capturing %s: %v\n", res.URL, err)
			} else {
				fmt.Printf(" [✓] Screenshot saved: %s\n", filename)
			}
		}

	case "version":
		fmt.Printf("NetVista v%s\n", version)
	default:
		fmt.Println("Expected 'scan' or 'version' subcommands")
		os.Exit(1)
	}
}
