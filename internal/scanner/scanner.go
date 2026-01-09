package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type PortScanner struct {
	Concurrency int
	Timeout     time.Duration
}

func NewPortScanner(concurrency int, timeout time.Duration) *PortScanner {
	return &PortScanner{
		Concurrency: concurrency,
		Timeout:     timeout,
	}
}

func (ps *PortScanner) Scan(ctx context.Context, host string, ports []int) []int {
	openPorts := make(chan int)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, ps.Concurrency)

	res := []int{}
	go func() {
		for _, port := range ports {
			wg.Add(1)
			go func(p int) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				address := fmt.Sprintf("%s:%d", host, p)
				conn, err := net.DialTimeout("tcp", address, ps.Timeout)
				if err == nil {
					conn.Close()
					openPorts <- p
				}
			}(port)
		}
		wg.Wait()
		close(openPorts)
	}()

	for p := range openPorts {
		res = append(res, p)
	}
	return res
}
