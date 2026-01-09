package domain

import (
	"time"
)

// Target represents a scanning endpoint.
type Target struct {
	URL      string
	IP       string
	Port     int
	Protocol string // http, https, etc.
}

// Metadata contains scan-time information about a host.
type Metadata struct {
	Title      string
	StatusCode int
	Technology []string
	Headers    map[string]string
	ContentLen int64
	Timestamp  time.Time
	Redirects  []string
}

// ScanResult aggregates all information gathered for a target.
type ScanResult struct {
	Target     Target
	Metadata   Metadata
	PHash      string
	Screenshot string // Path or identifier for the screenshot
	PHashScore uint64
	GroupID    string // Cluster/Group ID
	IsAlive    bool
	Error      string
}

// Config represents the application configuration.
type Config struct {
	Ports       string
	Concurrency int
	Timeout     time.Duration
	Proxy       string
	Headers     []string
	OutputPath  string
	FullPage    bool
	UserAgent   string
}
