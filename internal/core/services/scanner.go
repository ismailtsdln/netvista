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

// ScannerService coordinates the scanning process.
type ScannerService struct {
	prober    ports.Prober
	renderer  ports.Renderer
	analyzers []ports.Analyzer
	reporter  ports.Reporter
	config    domain.Config
	logger    *slog.Logger
}

// NewScannerService creates a new instance of the scanner service.
func NewScannerService(
	p ports.Prober,
	r ports.Renderer,
	a []ports.Analyzer,
	rep ports.Reporter,
	cfg domain.Config,
	logger *slog.Logger,
) *ScannerService {
	return &ScannerService{
		prober:    p,
		renderer:  r,
		analyzers: a,
		reporter:  rep,
		config:    cfg,
		logger:    logger,
	}
}

// Scan targets and generate a report.
func (s *ScannerService) Scan(ctx context.Context, targets []domain.Target) error {
	s.logger.Info("Starting advanced scan", "targets", len(targets), "concurrency", s.config.Concurrency)

	resultsChan := make(chan domain.ScanResult, len(targets))
	targetsChan := make(chan domain.Target, len(targets))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < s.config.Concurrency; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, targetsChan, resultsChan)
	}

	// Supply targets
	for _, t := range targets {
		targetsChan <- t
	}
	close(targetsChan)

	// Collect results
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var allResults []domain.ScanResult
	for res := range resultsChan {
		allResults = append(allResults, res)
	}

	// Generate final report
	if s.reporter != nil {
		if err := s.reporter.Write(allResults); err != nil {
			return fmt.Errorf("reporting failed: %w", err)
		}
	}

	return nil
}

func (s *ScannerService) worker(ctx context.Context, wg *sync.WaitGroup, targets <-chan domain.Target, results chan<- domain.ScanResult) {
	defer wg.Done()

	for t := range targets {
		select {
		case <-ctx.Done():
			return
		default:
			res := s.processTarget(ctx, t)
			results <- res
		}
	}
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
			if framework != "Static" {
				result.Metadata.Technology = append(result.Metadata.Technology, framework)
			}
		} else {
			s.logger.Warn("Render failed after retries", "url", result.Target.URL, "error", err)
		}
	}

	// 3. Analyze (Plugins)
	for _, analyzer := range s.analyzers {
		if err := analyzer.Analyze(ctx, &result); err != nil {
			s.logger.Warn("Analysis failed", "url", t.URL, "analyzer", analyzer.Name(), "error", err)
		}
	}

	return result
}
