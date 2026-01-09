package models

import "time"

type Target struct {
	Host     string
	Scheme   string
	Port     int
	URL      string
	IsAlive  bool
	PHash    string
	Metadata ResponseMetadata
}

type ResponseMetadata struct {
	Title         string
	StatusCode    int
	ContentLength int64
	Headers       map[string]string
	Technology    []string
	Timestamp     time.Time
}

type ScanResult struct {
	Target Target
	Image  []byte
	Error  string
}
