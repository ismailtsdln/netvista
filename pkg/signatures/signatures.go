package signatures

import (
	"embed"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed signatures.yaml
var DefaultSignaturesFS embed.FS

type FingerprintSig struct {
	Name   string            `yaml:"name"`
	Header map[string]string `yaml:"header"`
	Title  string            `yaml:"title"`
	Body   string            `yaml:"body"`
}

type TakeoverSig struct {
	Name        string `yaml:"name"`
	Fingerprint string `yaml:"fingerprint"`
}

type WafSig struct {
	Name   string            `yaml:"name"`
	Header map[string]string `yaml:"header"`
	Body   string            `yaml:"body"`
}

type Signatures struct {
	Fingerprints []FingerprintSig `yaml:"fingerprints"`
	Takeovers    []TakeoverSig    `yaml:"takeovers"`
	Wafs         []WafSig         `yaml:"wafs"`
}

func LoadSignatures(path string) (*Signatures, error) {
	var data []byte
	var err error

	if path != "" {
		data, err = os.ReadFile(path)
	}

	if err != nil || path == "" {
		// Fallback to embedded
		data, err = DefaultSignaturesFS.ReadFile("signatures.yaml")
		if err != nil {
			return nil, err
		}
	}

	var sigs Signatures
	err = yaml.Unmarshal(data, &sigs)
	if err != nil {
		return nil, err
	}

	return &sigs, nil
}
