package models

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
	StatusCode int               `json:"status_code"`
	Title      string            `json:"title"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Technology []string          `json:"technology"`
}

type ScanResult struct {
	Target Target
	Image  []byte
	Error  string
}
