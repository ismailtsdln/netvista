package engine

import (
	"bufio"
	"context"
	"os"
	"strings"
	"sync"

	"github.com/ismailtsdln/netvista/internal/prober"
	"github.com/ismailtsdln/netvista/pkg/models"
)

type Engine struct {
	Concurrency int
	Prober      *prober.Prober
}

func NewEngine(concurrency int, prob *prober.Prober) *Engine {
	return &Engine{
		Concurrency: concurrency,
		Prober:      prob,
	}
}

func (e *Engine) Run(ctx context.Context, targets []string) <-chan *models.Target {
	results := make(chan *models.Target)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, e.Concurrency)

	go func() {
		for _, target := range targets {
			wg.Add(1)
			go func(t string) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				res, err := e.Prober.Probe(ctx, t)
				if err == nil {
					results <- res
				}
			}(target)
		}
		wg.Wait()
		close(results)
	}()

	return results
}

func ReadTargetsFromStdin() []string {
	var targets []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			targets = append(targets, line)
		}
	}
	return targets
}
