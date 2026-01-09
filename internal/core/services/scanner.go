package services

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ismailtsdln/netvista/internal/core/domain"
	"github.com/ismailtsdln/netvista/internal/core/ports"
)

// ScannerService implements the core scanning logic.
type ScannerService struct {
	prober    ports.Prober
	renderer  ports.Renderer
	analyzers []ports.Analyzer
	reporter  ports.Reporter
	config    domain.Config
	logger    *slog.Logger
}

// NewScannerService creates a new ScannerService.
func NewScannerService(
	prober ports.Prober,
	renderer ports.Renderer,
	analyzers []ports.Analyzer,
	reporter ports.Reporter,
	config domain.Config,
	logger *slog.Logger,
) *ScannerService {
	return &ScannerService{
		prober:    prober,
		renderer:  renderer,
		analyzers: analyzers,
		reporter:  reporter,
		config:    config,
		logger:    logger,
	}
}

// Scan performs a scan on a list of targets.
func (s *ScannerService) Scan(ctx context.Context, targets []domain.Target) error {
	// Deduplicate targets by URL
	uniqueTargets := make(map[string]domain.Target)
	for _, t := range targets {
		uniqueTargets[t.URL] = t
	}

	targets = []domain.Target{}
	for _, t := range uniqueTargets {
		targets = append(targets, t)
	}

	s.logger.Info("Starting advanced scan", "targets", len(targets), "concurrency", s.config.Concurrency)

	var wg sync.WaitGroup
	results := make(chan domain.ScanResult, len(targets))
	workers := make(chan struct{}, s.config.Concurrency)

	for _, t := range targets {
		wg.Add(1)
		go func(target domain.Target) {
			defer wg.Done()
			workers <- struct{}{}
			defer func() { <-workers }()

			results <- s.processTarget(ctx, target)
		}(t)
	}

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	var scanResults []domain.ScanResult
	for res := range results {
		scanResults = append(scanResults, res)
		if res.Error != "" {
			s.logger.Warn("Scan result with error", "url", res.Target.URL, "error", res.Error)
		}
	}

	// Cluster results for reporting (simple grouping by domain or visual similarity)
	// This is now handled in the reporter adapter for v1 compatibility

	s.logger.Info("Scan completed, generating reports...")
	if s.reporter != nil {
		if err := s.reporter.Report(ctx, scanResults); err != nil {
			return fmt.Errorf("reporting failed: %w", err)
		}
	}

	return nil
}

func (s *ScannerService) processTarget(ctx context.Context, t domain.Target) domain.ScanResult {
	result := domain.ScanResult{
		Target: t,
	}

	// 1. Probe with Retry
	var metadata *domain.Metadata
	var resolvedURL string
	var err error

	for i := 0; i <= 2; i++ { // 3 attempts total
		metadata, resolvedURL, err = s.prober.Probe(ctx, t)
		if err == nil {
			break
		}
		if i < 2 {
			s.logger.Warn("Probe failed, retrying...", "url", t.URL, "attempt", i+1, "error", err)
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	if err != nil {
		result.Error = fmt.Sprintf("probe failed after retries: %v", err)
		result.IsAlive = false
		return result
	}
	result.Metadata = *metadata
	result.IsAlive = true
	result.Target.URL = resolvedURL

	// 2. Render (Screenshot) with optional retry
	if s.renderer != nil {
		var path, phash, framework string
		for i := 0; i <= 1; i++ { // 2 attempts total for rendering
			path, phash, framework, err = s.renderer.Render(ctx, result.Target)
			if err == nil {
				break
			}
			if i < 1 {
				s.logger.Warn("Render failed, retrying...", "url", result.Target.URL, "attempt", i+1, "error", err)
				time.Sleep(2 * time.Second)
			}
		}
		if err == nil {
			result.Screenshot = path
			result.PHash = phash
			if framework != "Static" && framework != "" {
				result.Metadata.Technology = append(result.Metadata.Technology, framework)
			}
		} else {
			s.logger.Warn("Render failed after retries", "url", result.Target.URL, "error", err)
		}
	}

	// 3. Analyze
	for _, analyzer := range s.analyzers {
		if err := analyzer.Analyze(ctx, &result); err != nil {
			s.logger.Warn("Analysis failed", "analyzer", analyzer.Name(), "url", t.URL, "error", err)
		}
	}

	return result
}
